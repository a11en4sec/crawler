package collect

// RuleTree 采集规则树
type RuleTree struct {
	Root  func() ([]*Request, error) // 根节点(执行入口),生成爬虫的种子网站
	Trunk map[string]*Rule           // 规则hash表,用于存储当前任务所有的规则
}

// Rule 采集规则节点
type Rule struct {
	ParseFunc func(*Context) (ParseResult, error) // 内容解析函数
}
