package master

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-micro.dev/v4/client"

	proto "github.com/a11en4sec/crawler/proto/crawler"
	"github.com/golang/protobuf/ptypes/empty"
)

func (m *Master) DeleteResource(ctx context.Context, spec *proto.ResourceSpec, empty *empty.Empty) error {
	// follow节点接受到请求,转发到leader
	if !m.IsLeader() && m.leaderID != "" && m.leaderID != m.ID {
		addr := getLeaderAddress(m.leaderID)
		_, err := m.forwardCli.DeleteResource(ctx, spec, client.WithAddress(addr))
		return err
	}

	// 加锁保障并发安全
	//m.rlock.Lock()
	//defer m.rlock.Unlock()

	// 取任务
	r, ok := m.resources[spec.Name]
	if !ok {
		return errors.New("no such task")
	}

	// etcd中删任务
	if _, err := m.etcdCli.Delete(context.Background(), getResourcePath(spec.Name)); err != nil {
		return err
	}

	// 程序内存中删除该任务:map中删除对应的key-value
	delete(m.resources, spec.Name)

	if r.AssignedNode != "" {
		nodeID, err := getNodeID(r.AssignedNode)
		if err != nil {
			return err
		}

		if ns, ok := m.workNodes[nodeID]; ok {
			ns.Payload -= 1
		}
	}
	return nil
}

func (m *Master) AddResource(ctx context.Context, req *proto.ResourceSpec, resp *proto.NodeSpec) error {
	// 如果不是 Leader，就获取 Leader 的地址,并完成请求的转发
	if !m.IsLeader() && m.leaderID != "" && m.leaderID != m.ID {

		// 获取leader的address
		addr := getLeaderAddress(m.leaderID)

		// 使用grpc client
		nodeSpec, err := m.forwardCli.AddResource(ctx, req, client.WithAddress(addr))
		resp.Id = nodeSpec.Id
		resp.Address = nodeSpec.Address
		return err
	}

	//m.rlock.Lock()
	//defer m.rlock.Unlock()

	fmt.Println("x:", len(m.workNodes))

	nodeSpec, err := m.addResources(&ResourceSpec{Name: req.Name})
	if nodeSpec != nil {
		resp.Id = nodeSpec.Node.Id
		resp.Address = nodeSpec.Node.Address
	}
	return err
}

func getLeaderAddress(address string) string {
	s := strings.Split(address, "-")
	if len(s) < 2 {
		return ""
	}
	return s[1]
}
