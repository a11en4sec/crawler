package proxy

import (
	"errors"
	"net/http"
	url "net/url"
	"sync/atomic"
)

type Func func(*http.Request) (*url.URL, error)

func RoundRobinProxySwitcher(ProxyURLs ...string) (Func, error) {
	if len(ProxyURLs) < 1 {
		return nil, errors.New("proxy URL list is empty")
	}

	urls := make([]*url.URL, len(ProxyURLs))

	for i, u := range ProxyURLs {
		parseURL, err := url.Parse(u)
		if err != nil {
			return nil, err
		}

		urls[i] = parseURL
	}

	return (&roundRobinSwitcher{urls, 0}).GetProxy, nil
}

type roundRobinSwitcher struct {
	proxyURLs []*url.URL
	index     uint32
}

// GetProxy 取余算法实现轮询调度
func (r *roundRobinSwitcher) GetProxy(pr *http.Request) (*url.URL, error) {
	if len(r.proxyURLs) == 0 {
		return nil, errors.New("empty proxy urls")
	}

	index := atomic.AddUint32(&r.index, 1) - 1
	u := r.proxyURLs[index%uint32(len(r.proxyURLs))]

	return u, nil
}
