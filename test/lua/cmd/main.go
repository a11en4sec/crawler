package main

import (
	"fmt"
	"github.com/yuin/gopher-lua"
)

func main() {
	L := lua.NewState()
	defer L.Close()
	//加载脚本
	err := L.DoFile("vardouble.lua")
	if err != nil {
		panic(err)
	}
	// 调用lua脚本内函数
	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("varDouble"), //指定要调用的函数名
		NRet:    1,                        // 指定返回值数量
		Protect: true,                     // 错误返回error
	}, lua.LNumber(15)) //支持多个参数
	if err != nil {
		panic(err)
	}
	//获取返回结果
	ret := L.Get(-1)
	//清理下，等待下次用
	L.Pop(1)

	//结果转下类型，方便输出
	res, ok := ret.(lua.LNumber)
	if !ok {
		panic("unexpected result")
	}
	fmt.Println(res.String())
}

// 输出结果：
// 30
