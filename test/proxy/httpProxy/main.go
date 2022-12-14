// 只能用于http协议请求代理，访问【https】的网站不能使用该代理
// 代理拿到req，自己往后端发起请求
package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	server := &http.Server{
		Addr:              ":8899",
		Handler:           http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handleHTTP(w, r) }),
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

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	// 代理服务器拿着req，自己向目标发起请求
	resp, err := http.DefaultTransport.RoundTrip(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	defer resp.Body.Close()

	// 1 把响应头copy回去
	copyHeader(w.Header(), resp.Header)
	// 2 回复响应码
	w.WriteHeader(resp.StatusCode)
	// 3 响应体copy到w中
	io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
