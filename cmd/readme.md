# 1 启动master
```shell
./main master --id=1 --http=:8081 --grpc=:9091 --pprof=:9981     
./main master --id=2 --http=:8082 --grpc=:9092 --pprof=:9982     
./main master --id=3 --http=:8083 --grpc=:9093 --pprof=:9983     

```
# 2 启动worker
```shell
./main worker --id=2 --http=:11801 --grpc=:11901 --pprof=:11001
./main worker --id=2 --http=:11802 --grpc=:11902 --pprof=:11002 
./main worker --id=2 --http=:11803 --grpc=:11903 --pprof=:11003
```
## 3 通过接口增加资源
```shell
curl -H "content-type: application/json" -d '{"id":"zjx","name": "task-test-4"}' http://localhost:8081/crawler/resource

{"id":"go.micro.server.worker-2", "Address":"192.168.0.107:9089"}
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