package main

import "fmt"

type Binary struct {
	uint64
}
type Stringer interface {
	String() string
}

func (i Binary) String() string {
	return "hello world"
}
func main() {
	a := Binary{54}
	// 实现接口的结构体:Binary接受对象
	b := Stringer(a)
	fmt.Printf("b.String:%s \n", b.String())
}
