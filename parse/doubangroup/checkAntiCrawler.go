package doubangroup

import (
	"fmt"
	"strings"
)

//nolint
func checkAntiCrawler(contents []byte) {
	if strings.Contains(string(contents), "有异常请求从你的") {
		fmt.Println("[---]:", "触发反爬")
	}
}
