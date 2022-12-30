# 1 安装
```shell
go get go-micro.dev/v4
```

# 2 验证etcd是否注册成功
```shell
docker exec etcd-gcr-v3.5.5 /bin/sh -c "/usr/local/bin/etcdctl get --prefix /"

/micro/registry/go.micro.server.worker/go.micro.server.worker-f7b923aa-cfa6-44ad-acb4-1b807c46f2c9
{"name":"go.micro.server.worker","version":"latest","metadata":null,"endpoints":[{"name":"Greeter.Hello","request":{"name":"Request","type":"Request","values":[{"name":"name","type":"string","values":null}]},"response":{"name":"Response","type":"Response","values":[{"name":"greeting","type":"string","values":null}]},"metadata":{"endpoint":"Greeter.Hello","handler":"rpc","method":"POST","path":"/greeter/hello"}}],"nodes":[{"id":"go.micro.server.worker-f7b923aa-cfa6-44ad-acb4-1b807c46f2c9","address":"192.168.50.199:9090","metadata":{"broker":"http","protocol":"grpc","registry":"etcd","server":"grpc","transport":"grpc"}}]}
```
