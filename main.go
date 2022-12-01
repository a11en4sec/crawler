package main

import (
	"fmt"
	"github.com/a11en4sec/crawler/collect"
	"github.com/a11en4sec/crawler/proxy"
	"log"
	"time"
)

func main() {

	proxyURLs := []string{"http://127.0.0.1:8889", "http://127.0.0.1:8888"}
	p, err := proxy.RoundRobinProxySwitcher(proxyURLs...)
	if err != nil {
		fmt.Println("RoundRobinProxySwitcher failed")
	}

	url := "https://book.douban.com/subject/1007305/"

	var f collect.Fetcher = collect.BrowserFetch{
		Timeout: 5000 * time.Millisecond,
		Proxy:   p,
	}
	body, err := f.Get(url)

	if err != nil {
		log.Println(err.Error())
	}

	fmt.Printf(string(body))

}
