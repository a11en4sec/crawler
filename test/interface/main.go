package main

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
	// todo 假如实现接口的结构体的接受对象
	b := Stringer(a)
	b.String()
}
