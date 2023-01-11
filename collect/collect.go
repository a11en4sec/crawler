package collect

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/a11en4sec/crawler/spider"

	"github.com/a11en4sec/crawler/extensions"
	"github.com/a11en4sec/crawler/proxy"
	"go.uber.org/zap"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type BaseFetch struct{}

func (BaseFetch) Get(req *spider.Request) ([]byte, error) {
	client := &http.Client{}
	ctx := context.Background()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, req.URL, nil)

	if err != nil {
		return nil, fmt.Errorf("get url failed:%w", err)
	}

	resp, err := client.Do(r)

	if err != nil {
		//fmt.Println(err)
		return nil, fmt.Errorf("get url failed:%w", err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			zap.L().Error("resp.body.close",
				zap.Error(err),
			)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		//fmt.Printf("Error status code:%d", resp.StatusCode)
		return nil, err
	}

	bodyReader := bufio.NewReader(resp.Body)
	e := DeterMinEncoding(bodyReader)
	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())

	return ioutil.ReadAll(utf8Reader)
}

type BrowserFetch struct {
	Timeout time.Duration
	Proxy   proxy.Func
	Logger  *zap.Logger
}

// Get 模拟浏览器访问
func (b BrowserFetch) Get(request *spider.Request) ([]byte, error) {
	client := &http.Client{
		Timeout: b.Timeout,
	}

	if b.Proxy != nil {
		transport := http.DefaultTransport.(*http.Transport)
		transport.Proxy = b.Proxy
		client.Transport = transport
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, request.URL, nil)

	if err != nil {
		return nil, fmt.Errorf("get url failed:%w", err)
	}

	if len(request.Task.Cookie) > 0 {
		req.Header.Set("Cookie", request.Task.Cookie)
	}

	req.Header.Set("User-Agent", extensions.GenerateRandomUA())

	resp, err := client.Do(req)

	if err != nil {
		// 统一在CreateWork 的 defer中处理
		//b.Logger.Error("fetch failed",
		//	zap.Error(err),
		//)
		return nil, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			b.Logger.Error("resp.body.close",
				zap.Error(err),
			)
		}
	}()

	bodyReader := bufio.NewReader(resp.Body)
	e := DeterMinEncoding(bodyReader)
	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())

	return ioutil.ReadAll(utf8Reader)
}

func DeterMinEncoding(r *bufio.Reader) encoding.Encoding {
	bys, err := r.Peek(1024)

	if err != nil {
		return unicode.UTF8
	}

	e, _, _ := charset.DetermineEncoding(bys, "")

	return e
}

//nolint
func checkAntiCrawler(closer io.ReadCloser) {
	// 检测是否被反爬拒绝--------
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(closer); err != nil {
		fmt.Println("[---]:", "read error")
	}

	respStr := buf.String()
	if strings.Contains(respStr, "有异常请求从你的") {
		fmt.Println("[---]:", "触发反爬")
	}
	// ----------------------

}
