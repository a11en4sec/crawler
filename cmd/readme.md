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