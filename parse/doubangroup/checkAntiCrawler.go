package doubangroup

import (
	"fmt"
	"strings"
)

func checkAntiCrawler(contents []byte) {
	// 检测是否被反扒拒绝--------
	//buf := new(bytes.Buffer)
	//buf.ReadFrom(closer)
	//respStr := buf.String()
	if strings.Contains(string(contents), "有异常请求从你的") {
		fmt.Println("[---]:", "触发反爬")
	}
	// ----------------------

}
