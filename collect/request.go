package collect

type Request struct {
	Url       string
	Cookie    string
	ParseFunc func([]byte, *Request) ParseResult
}

type ParseResult struct {
	Requesrts []*Request    // 用于进一步获取数据
	Items     []interface{} // 收到的的数据
}
