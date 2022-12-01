package proxy

import (
	"errors"
	"net/http"
	url "net/url"
	"sync/atomic"
)

type ProxyFunc func(*http.Request) (*url.URL, error)

func RoundRobinProxySwitcher(ProxyURLs ...string) (ProxyFunc, error) {
	if len(ProxyURLs) < 1 {
		return nil, errors.New("Proxy URL list is empty")
	}
	urls := make([]*url.URL, len(ProxyURLs))

	for i, u := range ProxyURLs {
		parseUrl, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		urls[i] = parseUrl
	}
	return (&roundRobinSwitcher{urls, 0}).GetProxy, nil
}

type roundRobinSwitcher struct {
	proxyURLs []*url.URL
	index     uint32
}

// GetProxy 取余算法实现轮询调度
func (r *roundRobinSwitcher) GetProxy(pr *http.Request) (*url.URL, error) {
	index := atomic.AddUint32(&r.index, 1) - 1
	u := r.proxyURLs[index%uint32(len(r.proxyURLs))]
	return u, nil
}
