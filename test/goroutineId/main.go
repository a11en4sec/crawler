package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"time"
)

func main() {
	go func() {
		gid := GetGid()
		fmt.Printf("child goruntine1 gid:%v \n", gid)
	}()
	go func() {
		gid := GetGid()
		fmt.Printf("child goruntine2 gid:%v \n", gid)
	}()
	go func() {
		gid := GetGid()
		fmt.Printf("child goruntine3 gid:%v \n", gid)
	}()
	go func() {
		gid := GetGid()
		fmt.Printf("child goruntine4 gid:%v \n", gid)
	}()
	go func() {
		gid := GetGid()
		fmt.Printf("child goruntine5 gid:%v \n", gid)
	}()
	gid := GetGid()
	fmt.Printf("main goruntine gid:%v \n", gid)
	time.Sleep(time.Second)
}
func GetGid() (gid uint64) {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		panic(err)
	}
	return n
}
