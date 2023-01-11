package main

import (
	"fmt"

	"go-micro.dev/v4/config"
	"go-micro.dev/v4/config/source/file"
)

func main() {

	// 导入数据
	err := config.Load(file.NewSource(
		file.WithPath("config.json"),
	))
	if err != nil {
		fmt.Println(err)
	}
	type Host struct {
		Address string `json:"address"`
		Port    int    `json:"port"`
	}

	var host Host
	// 获取hosts.database下的数据，并解析为host结构
	config.Get("hosts", "database").Scan(&host)

	fmt.Println(host)

	w, err := config.Watch("hosts", "database")
	if err != nil {
		fmt.Println(err)
	}

	// 等待配置文件更新
	v, err := w.Next()
	if err != nil {
		fmt.Println(err)
	}

	v.Scan(&host)
	fmt.Println(host)
}
