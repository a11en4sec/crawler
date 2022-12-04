package main

import (
	"fmt"
	"github.com/yuin/gopher-lua"
	"sync"
	"time"
)

//保存lua的LState的池子
type lStatePool struct {
	m     sync.Mutex
	saved []*lua.LState
}

// Get 获取一个LState
func (pl *lStatePool) Get() *lua.LState {
	pl.m.Lock()
	defer pl.m.Unlock()
	n := len(pl.saved)
	if n == 0 {
		return pl.New()
	}
	x := pl.saved[n-1]
	pl.saved = pl.saved[0 : n-1]
	return x
}

//New 新建一个LState
func (pl *lStatePool) New() *lua.LState {
	L := lua.NewState()
	// setting the L up here.
	// load scripts, set global variables, share channels, etc...
	//在这里我们可以做一些初始化

	return L
}

// Put 把Lstate对象放回到池中，方便下次使用
func (pl *lStatePool) Put(L *lua.LState) {
	pl.m.Lock()
	defer pl.m.Unlock()
	pl.saved = append(pl.saved, L)
}

// Shutdown 释放所有句柄
func (pl *lStatePool) Shutdown() {
	for _, L := range pl.saved {
		L.Close()
	}
}

// Global LState pool
var luaPool = &lStatePool{
	saved: make([]*lua.LState, 0, 4),
}

// MyWorker 协程内运行的任务
func MyWorker() {
	//通过pool获取一个LState
	L := luaPool.Get()

	//fmt.Printf("MyWorker:%#v \n", L)
	//任务执行完毕后，将LState放回pool
	defer luaPool.Put(L)
	// 这里可以用LState变量运行各种lua脚本任务
	// 例如 调用之前例子中的的varDouble函数

	if err := L.DoFile("vardouble.lua"); err != nil {
		panic(err)
	}

	err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("varDouble"), //指定要调用的函数名
		NRet:    1,                        // 指定返回值数量
		Protect: true,                     // 错误返回error
	}, lua.LNumber(25)) //这里支持多个参数

	if err != nil {
		panic(err) //仅供演示用，实际生产不推荐用panic
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
	fmt.Println("Last Result:", res.String())
}
func main() {
	defer luaPool.Shutdown()
	go MyWorker() // 启动一个协程
	go MyWorker() // 启动另外一个协程
	/* etc... */

	time.Sleep(5 * time.Second)

}
