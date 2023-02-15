package engine

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/a11en4sec/crawler/master"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/a11en4sec/crawler/spider"

	"go.uber.org/zap"
)

type Crawler struct {
	id          string
	out         chan spider.ParseResult
	Visited     map[string]bool // keys is md5(URL + method)
	VisitedLock sync.Mutex
	failures    map[string]*spider.Request // 失败请求id -> 失败请求
	failureLock sync.Mutex

	resources map[string]*master.ResourceSpec
	rlock     sync.Mutex

	etcdCli *clientv3.Client
	options
}

func (c *Crawler) Run(id string, cluster bool) {
	c.id = id
	// 单机模式
	if !cluster {
		// 从本地加载seed task
		c.handleSeeds()
	}

	// 集群模式(默认)
	go c.loadResource()  // 从etcd中加载资源(任务),并调用runTasks.
	go c.watchResource() // watch etcd的事件
	go c.Schedule()

	for i := 0; i < c.WorkCount; i++ {
		go c.CreateWork()
	}

	c.HandleResult()
}

func (c *Crawler) watchResource() {
	watch := c.etcdCli.Watch(context.Background(),
		master.RESOURCEPATH,
		clientv3.WithPrefix(),
		clientv3.WithPrevKV())
	for w := range watch {
		if w.Err() != nil {
			c.Logger.Error("watch resource failed", zap.Error(w.Err()))
			continue
		}
		if w.Canceled {
			c.Logger.Error("watch resource canceled")
			return
		}
		for _, ev := range w.Events {
			spec, err := master.Decode(ev.Kv.Value)
			if err != nil {
				c.Logger.Error("decode etcd value failed", zap.Error(err))
			}

			switch ev.Type {
			case clientv3.EventTypePut:
				// 删除事件
				spec, err := master.Decode(ev.Kv.Value)
				if err != nil {
					c.Logger.Error("decode etcd value failed", zap.Error(err))
				}

				if ev.IsCreate() {
					c.Logger.Info("receive create resource", zap.Any("spec", spec))

				} else if ev.IsModify() {
					c.Logger.Info("receive update resource", zap.Any("spec", spec))
				}
				c.rlock.Lock()
				c.runTasks(spec.Name)
				c.rlock.Unlock()

			case clientv3.EventTypeDelete:
				c.Logger.Info("receive delete resource", zap.Any("spec", spec))
				spec, err := master.Decode(ev.PrevKv.Value)
				if err != nil {
					c.Logger.Error("decode etcd value failed", zap.Error(err))
				}
				c.rlock.Lock()
				c.deleteTasks(spec.Name) // 把本地内存中的kv删除
				c.rlock.Unlock()
			}
		}
	}
}

func (c *Crawler) deleteTasks(taskName string) {
	t, ok := Store.Hash[taskName]
	if !ok {
		c.Logger.Error("can not find preset tasks", zap.String("task name", taskName))
		return
	}
	t.Closed = true
	delete(c.resources, taskName)
}

func getID(assignedNode string) string {
	s := strings.Split(assignedNode, "|")
	if len(s) < 2 {
		return ""
	}
	return s[0]
}

func (c *Crawler) loadResource() error {
	resp, err := c.etcdCli.Get(context.Background(), master.RESOURCEPATH, clientv3.WithPrefix(), clientv3.WithSerializable())
	if err != nil {
		return fmt.Errorf("etcd get failed")
	}

	resources := make(map[string]*master.ResourceSpec)
	for _, kv := range resp.Kvs {
		r, err := master.Decode(kv.Value)
		if err == nil && r != nil {
			id := getID(r.AssignedNode)
			if len(id) > 0 && c.id == id {
				resources[r.Name] = r
			}
		}
	}
	c.Logger.Info("leader init load resource", zap.Int("lenth", len(resources)))

	c.rlock.Lock()
	defer c.rlock.Unlock()
	c.resources = resources
	for _, r := range c.resources {
		c.runTasks(r.Name)
	}

	return nil
}

func (c *Crawler) runTasks(taskName string) {
	t, ok := Store.Hash[taskName]
	if !ok {
		c.Logger.Error("can not find preset tasks", zap.String("task name", taskName))
		return
	}
	t.Closed = false
	res, err := t.Rule.Root()

	if err != nil {
		c.Logger.Error("get root failed",
			zap.Error(err),
		)
		return
	}

	for _, req := range res {
		req.Task = t
	}
	c.scheduler.Push(res...)
}

func (c *Crawler) Schedule() {
	c.scheduler.Schedule()
}

func (c *Crawler) handleSeeds() {
	var reqQueue []*spider.Request

	for _, task := range c.Seeds {
		// 从全局store中取出Task
		t, ok := Store.Hash[task.Name]
		if !ok {
			c.Logger.Error("can not find preset tasks", zap.String("task name", task.Name))

			continue
		}
		// 1 将在main函数中初始化给seed的fetch，赋值给task
		//task.Fetcher = task.Fetcher
		// 2 将在main函数中初始化给seed的storage，赋值给task
		//task.Storage = task.Storage
		// 3 将在main函数中初始化给seed的limit，赋值给task
		//task.Limit = task.Limit

		task.Rule = t.Rule
		// 取出Task的根，根中存储的是种子request的列表
		rootReqs, err := task.Rule.Root()
		if err != nil {
			c.Logger.Error("get root failed",
				zap.Error(err),
			)

			continue
		}

		// 遍历，并且把task赋值给每一个req，后面处理爬回内容时，需要从task中获取一些信息
		for _, req := range rootReqs {
			req.Task = task
		}

		reqQueue = append(reqQueue, rootReqs...)
	}
	go c.scheduler.Push(reqQueue...)
}

func (c *Crawler) CreateWork() {
	defer func() {
		if err := recover(); err != nil {
			c.Logger.Error("worker panic",
				zap.Any("err", err),
				zap.String("stack", string(debug.Stack())))
		}
	}()

	for {
		req := c.scheduler.Pull() // 从workerChan中拉取一个req

		//判断r是否超过爬取的最高深度
		if err := req.Check(); err != nil {
			c.Logger.Error("check failed", zap.Error(err))

			continue
		}

		// 判断当前请求是否已被访问
		if !req.Task.Reload && c.HasVisited(req) {
			c.Logger.Debug("request has visited",
				zap.String("url:", req.URL),
			)

			continue
		}

		// 没有被访问，存到map中
		c.StoreVisited(req)

		//body, err := e.Fetcher.Get(req)
		body, err := req.Fetch()

		//fmt.Println("[++]", string(body))
		if err != nil {
			c.Logger.Error("can not fetch ",
				zap.Error(err),
				zap.String("url", req.URL),
			)

			// 请求失败，将请求放到错误map中
			c.SetFailure(req)

			continue
		}

		if len(body) < 6000 {
			c.Logger.Error("can't fetch ",
				zap.Int("length", len(body)),
				zap.String("url", req.URL),
			)

			// 请求失败，将请求放到错误map中
			c.SetFailure(req)

			continue
		}

		//result := req.ParseFunc(body, req)
		// //获取当前任务对应的规则 去处理fetch(req)回来的结果
		rule := req.Task.Rule.Trunk[req.RuleName]

		ctx := &spider.Context{
			Body: body,
			Req:  req,
		}

		// 处理fetch(req)回来的结果
		result, err := rule.ParseFunc(ctx)

		if err != nil {
			c.Logger.Error("ParseFunc failed ",
				zap.Error(err),
				zap.String("url", req.URL),
			)

			continue
		}

		// 解析结果里面新的url，继续爬
		if len(result.Requesrts) > 0 {
			go c.scheduler.Push(result.Requesrts...)
		}

		c.out <- result
	}
}

func (c *Crawler) HandleResult() {
	for result := range c.out {
		for _, item := range result.Items {
			c.Logger.Sugar().Info("get result ", item)

			switch d := item.(type) {
			case *spider.DataCell:
				//name := d.GetTaskName()
				//task := Store.Hash[name]
				//
				//if err := task.Storage.Save(d); err != nil {
				//	e.Logger.Error("存储失败")
				//}
				if err := d.Task.Storage.Save(d); err != nil {
					c.Logger.Error("")
				}
			}

			c.Logger.Sugar().Info("get result ", item)
		}
	}
}

func (c *Crawler) HasVisited(r *spider.Request) bool {
	c.VisitedLock.Lock()
	defer c.VisitedLock.Unlock()

	unique := r.Unique()

	return c.Visited[unique]
}

func (c *Crawler) StoreVisited(reqs ...*spider.Request) {
	c.VisitedLock.Lock()
	defer c.VisitedLock.Unlock()

	for _, r := range reqs {
		unique := r.Unique()
		c.Visited[unique] = true
	}
}

func (c *Crawler) SetFailure(req *spider.Request) {
	// 没有被重新爬取过,第一次fetch失败
	if !req.Task.Reload {
		c.VisitedLock.Lock()
		unique := req.Unique()
		// 从已爬取的map中删除该req
		delete(c.Visited, unique)
		c.VisitedLock.Unlock()
	}

	c.failureLock.Lock()
	defer c.failureLock.Unlock()

	// 失败队列中没有，说明是首次失败
	if _, ok := c.failures[req.Unique()]; !ok {
		// 首次失败时，再重新执行一次
		c.failures[req.Unique()] = req
		c.scheduler.Push(req)
	}
}
