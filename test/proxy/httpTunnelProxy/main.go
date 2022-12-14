// 可以代理http和https请求,中间不拆包
// [client]<--->(clientConn)[Proxy](destConn)<----->[server]
// 数据在客户端和服务端之间来回拷贝
package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func main() {
	server := &http.Server{
		Addr: ":8889",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 在 HTTP 隧道技术中，客户端会在第一次连接代理服务器时给代理服务器发送一个指令，通常是一个 HTTP 请求。
			// 这里我们可以将 HTTP 请求头中的 method 设置为 CONNECT
			// 代理服务器收到该指令后，将与目标服务器建立 TCP 连接
			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
		TLSConfig:         nil,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
		BaseContext:       nil,
		ConnContext:       nil,
	}

	log.Fatal(server.ListenAndServe())
}

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	// 客户端与代理服务器之间的底层 TCP 连接
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	// [client]<--->(clientConn)[Proxy](destConn)<----->[server]
	// 数据在客户端和服务端之间来回拷贝
	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)

	//var buf1 []byte
	//var buf2 []byte
	//
	//go copyBuffer(destConn, clientConn, buf1)
	//go copyBuffer(clientConn, destConn, buf2)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func copyBuffer(dst io.Writer, src io.Reader, buf []byte) (int64, error) {
	if len(buf) == 0 {
		buf = make([]byte, 32*1024)
	}
	var written int64
	for {
		nr, rerr := src.Read(buf)
		if rerr != nil && rerr != io.EOF && rerr != context.Canceled {
			log.Printf("httputil: ReverseProxy read error during body copy: %v", rerr)
		}
		if nr > 0 {
			nw, werr := dst.Write(buf[:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if werr != nil {
				return written, werr
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if rerr != nil {
			if rerr == io.EOF {
				rerr = nil
			}
			return written, rerr
		}
	}
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
