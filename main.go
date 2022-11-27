package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
	"regexp"
)

//var headerRe = regexp.MustCompile(`<div class="news_li"[\s\S]*?<h2>[\s\S]*?<a.*?target="_blank">([\s\S]*?)</a>`)
var headerRe = regexp.MustCompile(`<div class="small_cardcontent__BTALp"[\s\S]*?<h2>([\s\S]*?)</h2>`)

func main1() {
	url := "https://www.thepaper.cn/"

	body, err := Fetch(url)

	if err != nil {
		fmt.Printf("read content failed:%v", err)
		return
	}

	matches := headerRe.FindAllSubmatch(body, -1)

	for _, m := range matches {
		fmt.Println("fetch card news:", string(m[1]))
	}
}

func Fetch(url string) ([]byte, error) {

	resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error status code :%d", resp.StatusCode)
	}

	bodyReader := bufio.NewReader(resp.Body)
	e := DeterminEncoding(bodyReader)

	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())

	return ioutil.ReadAll(utf8Reader)

}

// DeterminEncoding 检测网页的字符编码并将文本统一转换为 UTF-8 格式
func DeterminEncoding(r *bufio.Reader) encoding.Encoding {
	bytes, err := r.Peek(1024)

	if err != nil {
		fmt.Printf("fetch error :%v", err)
		return unicode.UTF8
	}

	e, _, _ := charset.DetermineEncoding(bytes, "")

	return e
}

// tag v0.0.6
func main2() {
	url := "https://www.thepaper.cn/"
	body, err := Fetch(url)

	if err != nil {
		fmt.Printf("read content failed:%v", err)
		return
	}
	doc, err := htmlquery.Parse(bytes.NewReader(body))
	if err != nil {
		fmt.Printf("htmlquery.Parse failed:%v", err)
	}
	nodes := htmlquery.Find(doc, `//div[@class="small_cardcontent__BTALp"]/div/a[@target="_blank"]/h2`)

	for _, node := range nodes {
		fmt.Println("fetch card ", node.FirstChild.Data)
	}
}

// tag v0.0.9
func main() {
	url := "https://www.thepaper.cn/"
	body, err := Fetch(url)

	if err != nil {
		fmt.Printf("read content failed:%v", err)
		return
	}

	// 加载HTML文档
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		fmt.Printf("read content failed:%v", err)
	}

	doc.Find("div.small_cardcontent__BTALp div a[target=_blank] h2").Each(func(i int, s *goquery.Selection) {
		// 获取匹配标签中的文本
		title := s.Text()
		fmt.Printf("Review %d: %s\n", i, title)
	})
}
