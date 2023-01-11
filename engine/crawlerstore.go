package engine

import (
	"github.com/a11en4sec/crawler/spider"
)

type CrawlerStore struct {
	list []*spider.Task
	Hash map[string]*spider.Task
}

func (c *CrawlerStore) Add(task *spider.Task) {
	c.Hash[task.Name] = task
	c.list = append(c.list, task)
}

// Store 全局爬虫任务实例
var Store = &CrawlerStore{
	list: []*spider.Task{},
	Hash: map[string]*spider.Task{},
}

func GetFields(taskName string, ruleName string) []string {
	return Store.Hash[taskName].Rule.Trunk[ruleName].ItemFields
}
