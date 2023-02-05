package master

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	proto "github.com/a11en4sec/crawler/proto/crawler"

	"github.com/bwmarrin/snowflake"

	"github.com/a11en4sec/crawler/cmd/worker"

	"go.uber.org/zap"

	"go.etcd.io/etcd/client/v3/concurrency"

	clientv3 "go.etcd.io/etcd/client/v3"

	"go-micro.dev/v4/registry"
)

const (
	RESOURCEPATH = "/resources"
)

type Master struct {
	ID         string
	ready      int32
	leaderID   string
	workNodes  map[string]*NodeSpec     // master中存储所有的work节点
	resources  map[string]*ResourceSpec // master中存储的资源
	IDGen      *snowflake.Node
	etcdCli    *clientv3.Client
	forwardCli proto.CrawlerMasterService
	options
}

// SetForwardCli 将生成的 GRPC client 注入到了 Master 结构体中
func (m *Master) SetForwardCli(forwardCli proto.CrawlerMasterService) {
	m.forwardCli = forwardCli
}

type Command int

const (
	MSGADD Command = iota
	MSGDELETE
)

type Message struct {
	Cmd   Command
	Specs []*ResourceSpec
}

type ResourceSpec struct {
	ID           string
	Name         string
	AssignedNode string
	CreationTime int64
}

type NodeSpec struct {
	Node    *registry.Node
	Payload int
}

func New(id string, opts ...Option) (*Master, error) {
	m := &Master{}

	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	m.resources = make(map[string]*ResourceSpec)
	m.options = options

	node, err := snowflake.NewNode(1)
	if err != nil {
		return nil, err
	}
	m.IDGen = node

	ipv4, err := getLocalIP()
	if err != nil {
		return nil, err
	}
	m.ID = genMasterID(id, ipv4, m.GRPCAddress)
	m.logger.Sugar().Debugln("master_id:", m.ID)

	// master中的etcd cli
	endpoints := []string{m.registryURL}
	cli, err := clientv3.New(clientv3.Config{Endpoints: endpoints})
	if err != nil {
		return nil, err
	}
	m.etcdCli = cli

	// 从etcd中拉一下信息
	m.updateWorkNodes() // 更新WorkNodes
	m.AddSeed()

	// master参与Leader竞选
	go m.Campaign()

	go m.HandleMsg()

	return m, nil
}

func genMasterID(id string, ipv4 string, GRPCAddress string) string {
	return "master" + id + "-" + ipv4 + GRPCAddress
}

func (m *Master) IsLeader() bool {
	return atomic.LoadInt32(&m.ready) != 0
}

func (m *Master) Campaign() {
	// 1 创建一个与 etcd 服务端带租约的会话
	s, err := concurrency.NewSession(m.etcdCli, concurrency.WithTTL(5))
	if err != nil {
		fmt.Println("NewSession", "error", "err", err)
	}
	defer func() {
		err := s.Close()
		if err != nil {
			fmt.Println("Campaign etcd session close ", "error", "err", err)
		}
	}()

	// 2 创建一个新的etcd选举对象，抢占到该 Key 的 Master 将变为 Leader
	e := concurrency.NewElection(s, "/crawler/election")
	leaderCh := make(chan error)

	// 让当前master进入Leader选举
	go m.elect(e, leaderCh)

	// 监听 Leader 的变化，当 Leader 状态发生变化时，会将当前 Leader 的信息发送到通道中
	leaderChange := e.Observe(context.Background())

	// 初始化时首先[堵塞]读取了一次 e.Observe 返回的通道信息
	// 只有成功收到 e.Observe 返回的消息，才意味着集群中已经存在 Leader，表示集群完成了选举。
	select {
	case resp := <-leaderChange:
		m.logger.Info("watch leader change", zap.String("leader:", string(resp.Kvs[0].Value)))
		// 所有 Master 节点要在 Leader 发生变更时，将当前最新的 Leader 地址保存到 leaderID 中
		m.leaderID = string(resp.Kvs[0].Value)

	}
	workerNodeChange := m.WatchWorker()

	for {
		select {
		// 监听当前 Master 是否当上了 Leader
		case err := <-leaderCh:
			if err != nil {
				m.logger.Error("leader elect failed", zap.Error(err))
				go m.elect(e, leaderCh)
			} else {
				m.logger.Info("master start change to leader")
				m.leaderID = m.ID
				if !m.IsLeader() {
					if err := m.BecomeLeader(); err != nil {
						m.logger.Error("BecomeLeader failed", zap.Error(err))
					}
				}
			}
			// 监听当前集群中 Leader 是否发生了变化
		case resp := <-leaderChange:
			if len(resp.Kvs) > 0 {
				m.logger.Info("watch leader change", zap.String("leader:", string(resp.Kvs[0].Value)))
			}
		case resp := <-workerNodeChange:
			m.logger.Info("watch worker change", zap.Any("worker:", resp))
			m.updateWorkNodes()
			if err := m.loadResource(); err != nil {
				m.logger.Error("loadResource failed:%w", zap.Error(err))
			}
			// 重新分配资源
			m.reAssign()

		case <-time.After(20 * time.Second):
			rsp, err := e.Leader(context.Background())
			if err != nil {
				m.logger.Info("get Leader failed", zap.Error(err))
				if errors.Is(err, concurrency.ErrElectionNoLeader) {
					go m.elect(e, leaderCh)
				}
			}
			if rsp != nil && len(rsp.Kvs) > 0 {
				m.logger.Debug("get Leader", zap.String("value", string(rsp.Kvs[0].Value)))
				if m.IsLeader() && m.ID != string(rsp.Kvs[0].Value) {
					//当前已不再是leader
					atomic.StoreInt32(&m.ready, 0)
				}
			}
		}
	}
}

func (m *Master) elect(e *concurrency.Election, ch chan error) {
	// 堵塞直到选取成功
	err := e.Campaign(context.Background(), m.ID)
	ch <- err
}

// WatchWorker Master 对 Worker 的服务发现
func (m *Master) WatchWorker() chan *registry.Result {
	// 服务发现使用 micro 提供的 registry 功能
	watch, err := m.registry.Watch(registry.WatchService(worker.ServiceName))
	if err != nil {
		panic(err)
	}
	ch := make(chan *registry.Result)
	go func() {
		for {
			// 持续获取值
			res, err := watch.Next()
			if err != nil {
				m.logger.Info("watch worker service failed", zap.Error(err))
				continue
			}
			// 发送到通道
			ch <- res
		}
	}()
	return ch
}

func (m *Master) BecomeLeader() error {
	// 成为leader后，更新下worker
	m.updateWorkNodes()

	// 成为leader后，重新加载资源
	if err := m.loadResource(); err != nil {
		return fmt.Errorf("loadResource failed:%w", err)
	}

	// 重新分配资源
	m.reAssign()

	// 跟新状态
	atomic.StoreInt32(&m.ready, 1)

	return nil
}

func (m *Master) reAssign() {
	rs := make([]*ResourceSpec, 0, len(m.resources))
	// 遍历资源
	// 发现有资源还没有分配节点时，将再次尝试将资源分配到 Worker 中。
	// 如果发现资源都已经分配给了对应的 Worker，它就会查看当前节点是否存活。
	// 如果当前节点已经不存在了，就将该资源分配给其他的节点。
	for _, r := range m.resources {
		if r.AssignedNode == "" {
			rs = append(rs, r)
			continue
		}

		id, err := getNodeID(r.AssignedNode)

		if err != nil {
			m.logger.Error("get nodeid failed", zap.Error(err))
		}

		if _, ok := m.workNodes[id]; !ok {
			rs = append(rs, r)
		}
	}

	m.AddResources(rs)
}

func getNodeID(assigned string) (string, error) {
	node := strings.Split(assigned, "|")
	if len(node) < 2 {
		return "", errors.New("")
	}
	id := node[0]
	return id, nil
}

//func (m *Master) DeleteResource(ctx context.Context, spec *proto.ResourceSpec, empty *empty.Empty) error {
//	r, ok := m.resources[spec.Name]
//
//	if !ok {
//		return errors.New("no such task")
//	}
//
//	if _, err := m.etcdCli.Delete(context.Background(), getResourcePath(spec.Name)); err != nil {
//		return err
//	}
//
//	if r.AssignedNode != "" {
//		nodeID, err := getNodeID(r.AssignedNode)
//		if err != nil {
//			return err
//		}
//
//		if ns, ok := m.workNodes[nodeID]; ok {
//			ns.Payload -= 1
//		}
//	}
//	return nil
//}

func (m *Master) AddResources(rs []*ResourceSpec) {
	for _, r := range rs {
		_, err := m.addResources(r)
		if err != nil {
			m.logger.Error("AddResources", zap.Error(err))
		}
	}
}

//func (m *Master) AddResource(ctx context.Context, req *proto.ResourceSpec, resp *proto.NodeSpec) error {
//	nodeSpec, err := m.addResources(&ResourceSpec{Name: req.Name})
//	if nodeSpec != nil {
//		resp.Id = nodeSpec.Node.Id
//		resp.Address = nodeSpec.Node.Address
//	}
//	return err
//}

func (m *Master) updateWorkNodes() {
	services, err := m.registry.GetService(worker.ServiceName)
	if err != nil {
		m.logger.Error("get service", zap.Error(err))
	}

	nodes := make(map[string]*NodeSpec)
	if len(services) > 0 {
		for _, spec := range services[0].Nodes {
			nodes[spec.Id] = &NodeSpec{
				Node: spec,
			}
		}
	}

	added, deleted, changed := workNodeDiff(m.workNodes, nodes)
	m.logger.Sugar().Info("worker joined: ", added, ", leaved: ", deleted, ", changed: ", changed)

	m.workNodes = nodes

}

func workNodeDiff(old map[string]*NodeSpec, new map[string]*NodeSpec) ([]string, []string, []string) {
	added := make([]string, 0)
	deleted := make([]string, 0)
	changed := make([]string, 0)
	for k, v := range new {
		if ov, ok := old[k]; ok {
			if !reflect.DeepEqual(v.Node, ov.Node) {
				changed = append(changed, k)
			}
		} else {
			added = append(added, k)
		}
	}
	for k := range old {
		if _, ok := new[k]; !ok {
			deleted = append(deleted, k)
		}
	}
	return added, deleted, changed
}

// 获取本机网卡IP
func getLocalIP() (string, error) {
	var (
		addrs []net.Addr
		err   error
	)
	// 获取所有网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return "", err
	}
	// 取第一个非lo的网卡IP
	for _, addr := range addrs {
		if ipNet, isIpNet := addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", errors.New("no local ip")
}

func getResourcePath(name string) string {
	return fmt.Sprintf("%s/%s", RESOURCEPATH, name)
}

func encode(s *ResourceSpec) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func decode(ds []byte) (*ResourceSpec, error) {
	var s *ResourceSpec
	err := json.Unmarshal(ds, &s)
	return s, err
}

// addResources 添加一个资源
func (m *Master) addResources(r *ResourceSpec) (*NodeSpec, error) {
	// 雪花算法生成一个id
	r.ID = m.IDGen.Generate().String()
	// 分配资源
	ns, err := m.Assign(r)
	if err != nil {
		m.logger.Error("assign failed", zap.Error(err))
		return nil, err
	}

	// 没有找到worker node
	if ns.Node == nil {
		m.logger.Error("no node to assgin")
		return nil, err
	}

	r.AssignedNode = ns.Node.Id + "|" + ns.Node.Address
	r.CreationTime = time.Now().UnixNano()
	m.logger.Debug("add resource", zap.Any("specs", r))

	// 写入etcd中
	_, err = m.etcdCli.Put(context.Background(), getResourcePath(r.Name), encode(r))
	if err != nil {
		m.logger.Error("put etcd failed", zap.Error(err))
		return nil, err
	}

	// master中记录下,方便查找
	m.resources[r.Name] = r
	ns.Payload++

	return ns, nil

}

func (m *Master) loadResource() error {
	resp, err := m.etcdCli.Get(context.Background(), RESOURCEPATH, clientv3.WithPrefix(), clientv3.WithSerializable())
	if err != nil {
		return fmt.Errorf("etcd get failed")
	}

	resources := make(map[string]*ResourceSpec)
	for _, kv := range resp.Kvs {
		r, err := decode(kv.Value)
		if err == nil && r != nil {
			resources[r.Name] = r
		}
	}
	m.logger.Info("leader init load resource", zap.Int("lenth", len(m.resources)))
	m.resources = resources

	for _, r := range m.resources {
		// 资源没有分配worker node
		if r.AssignedNode != "" {
			id, err := getNodeID(r.AssignedNode)
			if err != nil {
				m.logger.Error("getNodeID failed", zap.Error(err))
			}
			if node, ok := m.workNodes[id]; ok {
				node.Payload++
			}
		}
	}

	return nil
}

// Assign 把资源分配给worker节点
func (m *Master) Assign(*ResourceSpec) (*NodeSpec, error) {

	candidates := make([]*NodeSpec, 0, len(m.workNodes))

	for _, node := range m.workNodes {
		candidates = append(candidates, node)
	}

	//  找到最低的负载
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Payload < candidates[j].Payload
	})

	if len(candidates) > 0 {
		return candidates[0], nil
	}

	// 遍历map，等于随机分配一个
	//for _, n := range m.workNodes {
	//	return n, nil
	//}

	return nil, errors.New("no worker nodes")
}

// AddSeed 写入etcd，并更新master
func (m *Master) AddSeed() {
	rs := make([]*ResourceSpec, 0, len(m.Seeds))

	// 根据种子任务的Name,从etcd中取
	for _, seed := range m.Seeds {
		resp, err := m.etcdCli.Get(context.Background(), getResourcePath(seed.Name), clientv3.WithSerializable(), clientv3.WithPrefix())
		if err != nil {
			m.logger.Error("etcd get faiiled", zap.Error(err))
			continue
		}
		// 没有取到，说明etcd中存的是空
		if len(resp.Kvs) == 0 {
			r := &ResourceSpec{
				Name: seed.Name,
			}
			rs = append(rs, r)
		}
	}

	m.AddResources(rs)
}

func (m *Master) HandleMsg() {
	msgCh := make(chan *Message)

	select {
	case msg := <-msgCh:
		switch msg.Cmd {
		case MSGADD:
			m.AddResources(msg.Specs)
		case MSGDELETE:

		}
	}

}
