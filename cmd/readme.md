# 1 启动master
```shell
./main master --id=1 --http=:8081 --grpc=:9091 --pprof=:9981     
./main master --id=2 --http=:8082 --grpc=:9092 --pprof=:9982     
./main master --id=4 --http=:8084 --grpc=:9094 --pprof=:9984     

```
# 2 启动worker
```shell
./main worker --id=1 --http=:11801 --grpc=:11901 --pprof=:11001
./main worker --id=2 --http=:11802 --grpc=:11902 --pprof=:11002 
./main worker --id=3 --http=:11803 --grpc=:11903 --pprof=:11003
```
## 3 通过接口增加资源
### 3.1 直接请求leader节点添加资源
```shell
curl -H "content-type: application/json" -d '{"id":"zjx","name": "task-test-4"}' http://localhost:8081/crawler/resource

{"id":"go.micro.server.worker-2", "Address":"192.168.0.107:9089"}
```

### 3.2 直接请求follow节点添加资源(会转发)
角色是follow的master收到请求后，会通过grpc client,向leader转发请求,过程如下:

```shell

curl  --request POST 'http://localhost:8082/crawler/resource' --header 'Content-Type: application/json' --data '{"id":"zjx","name": "task-forward"}' 
{"id":"go.micro.server.worker-1","Address":"192.168.0.105:9090"}


// follow节点日志
"msg":"receive request","method":"CrawlerMaster.AddResource","Service":"go.micro.server.master","request param:":{"id":"zjx","name":"task-forward"}}

// leader节点日志
"msg":"receive request","method":"CrawlerMaster.AddResource","Service":"go.micro.server.master","request param:":{"id":"zjx","name":"task-forward"}}
"msg":"add resource","specs":{"ID":"1622019481540759552","Name":"task-forward","AssignedNode":"go.micro.server.worker-1|192.168.50.199:11901","CreationTime":1675554554713968000}}
```
### 原理
1. Master中有成员forwardCli
```go
type Master struct {
	ID         string
	ready      int32
	leaderID   string
	workNodes  map[string]*NodeSpec     // master中存储所有的work节点
	resources  map[string]*ResourceSpec // master中存储的资源
	IDGen      *snowflake.Node
	etcdCli    *clientv3.Client
	forwardCli proto.CrawlerMasterService // 接口
	options
}
```
```go
// Client API for CrawlerMaster service

type CrawlerMasterService interface {
	AddResource(ctx context.Context, in *ResourceSpec, opts ...client.CallOption) (*NodeSpec, error)
	DeleteResource(ctx context.Context, in *ResourceSpec, opts ...client.CallOption) (*emptypb.Empty, error)
}
```
2. 初始化时导入 micro GRPC client 的插件实现的。SetForwardCli 方法将生成的 GRPC client 注入到了 Master 结构体中。
```go

import (
  grpccli "github.com/go-micro/plugins/v4/client/grpc"
)

func RunGRPCServer(m *master.Master, logger *zap.Logger, reg registry.Registry, cfg ServerConfig) {
  service := micro.NewService(
    ...
    micro.Client(grpccli.NewClient()), 
  )

  cl := proto.NewCrawlerMasterService(cfg.Name, service.Client())
  m.SetForwardCli(cl)
}
```
```go
// SetForwardCli 将生成的 GRPC client 注入到了 Master 结构体中
func (m *Master) SetForwardCli(forwardCli proto.CrawlerMasterService) {
	m.forwardCli = forwardCli
}
````

3. Follow转发请求至Leader

数据流：
```shell
User http request ---> Follow http Server
                           |
                           v
                        Follow gRPC Server  ----> Leader gRPC server
```

转发逻辑：
```go
func (m *Master) AddResource(ctx context.Context, req *proto.ResourceSpec, resp *proto.NodeSpec) error {
	// 如果自己不是 Leader，获取 Leader 的地址,并完成请求的转发
	if !m.IsLeader() && m.leaderID != "" && m.leaderID != m.ID {
		// 获取leader的address
		addr := getLeaderAddress(m.leaderID)

		// 使用grpc client,向Leader发起调用
		nodeSpec, err := m.forwardCli.AddResource(ctx, req, client.WithAddress(addr))
		resp.Id = nodeSpec.Id
		resp.Address = nodeSpec.Address
		return err
	}

	nodeSpec, err := m.addResources(&ResourceSpec{Name: req.Name})
	if nodeSpec != nil {
		resp.Id = nodeSpec.Node.Id
		resp.Address = nodeSpec.Node.Address
	}
	return err
}
```

```

# 信息
## 增加worker时
```json
{
	"level": "INFO",
	"ts": "2022-12-12T16:55:42.798+0800",
	"logger": "master",
	"caller": "master/master.go:117",
	"msg": "watch worker change",
	"worker:": {
		"Action": "create",
		"Service": {
			"name": "go.micro.server.worker",
			"version": "latest",
			"metadata": null,
			"endpoints": [{
				"name": "Greeter.Hello",
				"request": {
					"name": "Request",
					"type": "Request",
					"values": [{
						"name": "name",
						"type": "string",
						"values": null
					}]
				},
				"response": {
					"name": "Response",
					"type": "Response",
					"values": [{
						"name": "greeting",
						"type": "string",
						"values": null
					}]
				},
				"metadata": {
					"endpoint": "Greeter.Hello",
					"handler": "rpc",
					"method": "POST",
					"path": "/greeter/hello"
				}
			}],
			"nodes": [{
				"id": "go.micro.server.worker-2",
				"address": "192.168.0.107:9089",
				"metadata": {
					"broker": "http",
					"protocol": "grpc",
					"registry": "etcd",
					"server": "grpc",
					"transport": "grpc"
				}
			}]
		}
	}
}
```

## 删除worker时
```json
{
	"Action": "delete",
	"Service": {
		"name": "go.micro.server.worker",
		"version": "latest",
		"metadata": null,
		"endpoints": [{
			"name": "Greeter.Hello",
			"request": {
				"name": "Request",
				"type": "Request",
				"values": [{
					"name": "name",
					"type": "string",
					"values": null
				}]
			},
			"response": {
				"name": "Response",
				"type": "Response",
				"values": [{
					"name": "greeting",
					"type": "string",
					"values": null
				}]
			},
			"metadata": {
				"endpoint": "Greeter.Hello",
				"handler": "rpc",
				"method": "POST",
				"path": "/greeter/hello"
			}
		}],
		"nodes": [{
			"id": "go.micro.server.worker-3",
			"address": "192.168.50.199:11003",
			"metadata": {
				"broker": "http",
				"protocol": "grpc",
				"registry": "etcd",
				"server": "grpc",
				"transport": "grpc"
			}
		}]
	}
}
}
```

## etcd中的kv
```json
/micro/registry/go.micro.server.master/go.micro.server.master-1
{"name":"go.micro.server.master","version":"latest","metadata":null,"endpoints":[{"name":"Greeter.Hello","request":{"name":"Request","type":"Request","values":[{"name":"name","type":"string","values":null}]},"response":{"name":"Response","type":"Response","values":[{"name":"greeting","type":"string","values":null}]},"metadata":{"endpoint":"Greeter.Hello","handler":"rpc","method":"POST","path":"/greeter/hello"}}],"nodes":[{"id":"go.micro.server.master-1","address":"192.168.50.199:9091","metadata":{"broker":"http","protocol":"grpc","registry":"etcd","server":"grpc","transport":"grpc"}}]}

/micro/registry/go.micro.server.master/go.micro.server.master-3
{"name":"go.micro.server.master","version":"latest","metadata":null,"endpoints":[{"name":"Greeter.Hello","request":{"name":"Request","type":"Request","values":[{"name":"name","type":"string","values":null}]},"response":{"name":"Response","type":"Response","values":[{"name":"greeting","type":"string","values":null}]},"metadata":{"endpoint":"Greeter.Hello","handler":"rpc","method":"POST","path":"/greeter/hello"}}],"nodes":[{"id":"go.micro.server.master-3","address":"192.168.50.199:9093","metadata":{"broker":"http","protocol":"grpc","registry":"etcd","server":"grpc","transport":"grpc"}}]}

/micro/registry/go.micro.server.worker/go.micro.server.worker-2
{"name":"go.micro.server.worker","version":"latest","metadata":null,"endpoints":[{"name":"Greeter.Hello","request":{"name":"Request","type":"Request","values":[{"name":"name","type":"string","values":null}]},"response":{"name":"Response","type":"Response","values":[{"name":"greeting","type":"string","values":null}]},"metadata":{"endpoint":"Greeter.Hello","handler":"rpc","method":"POST","path":"/greeter/hello"}}],"nodes":[{"id":"go.micro.server.worker-2","address":"192.168.50.199:9089","metadata":{"broker":"http","protocol":"grpc","registry":"etcd","server":"grpc","transport":"grpc"}}]}

/micro/registry/go.micro.server.worker/go.micro.server.worker-3
{"name":"go.micro.server.worker","version":"latest","metadata":null,"endpoints":[{"name":"Greeter.Hello","request":{"name":"Request","type":"Request","values":[{"name":"name","type":"string","values":null}]},"response":{"name":"Response","type":"Response","values":[{"name":"greeting","type":"string","values":null}]},"metadata":{"endpoint":"Greeter.Hello","handler":"rpc","method":"POST","path":"/greeter/hello"}}],"nodes":[{"id":"go.micro.server.worker-3","address":"192.168.50.199:11003","metadata":{"broker":"http","protocol":"grpc","registry":"etcd","server":"grpc","transport":"grpc"}}]}

# 任务1
/resources/douban_book_list
{"ID":"1621665639791857664","Name":"douban_book_list","AssignedNode":"go.micro.server.worker-2|192.168.50.199:9089","CreationTime":1675470192274320000}

# 任务2
/resources/election/3f358619bf221967
master1-192.168.50.199:9091

# master选出来的leader
/resources/election/3f358619bf221984
master3-192.168.50.199:9093
```