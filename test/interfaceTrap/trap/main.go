package main

type adder interface {
	Add()
}
type M struct{}

func (m *M) Add() {

}
func main() {
	var m adder = &M{}
	m.Add()
}

// Trap
/*
func main() {
	var m adder = M{}  // 编译不通过，因为实现接口的结构体是*M,而不是M
	m.Add()
}

*/
