package engine

import "github.com/a11en4sec/crawler/collect"

type CrawlerStore struct {
	list []*collect.Task
	Hash map[string]*collect.Task
}

func (c *CrawlerStore) Add(task *collect.Task) {
	c.Hash[task.Name] = task
	c.list = append(c.list, task)
}

// Store 全局爬虫任务实例
var Store = &CrawlerStore{
	list: []*collect.Task{},
	Hash: map[string]*collect.Task{},
}

func GetFields(taskName string, ruleName string) []string {
	return Store.Hash[taskName].Rule.Trunk[ruleName].ItemFields
}
