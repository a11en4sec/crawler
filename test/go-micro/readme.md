```shell

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

go-micro 在此基础上进行了扩展，我们需要下   载 protoc-gen-micro 插件来生成 micro 适用的协议文件。这个插件的版本需要和我们使用的 go-micro 版本相同。目前，最新的 go-micro 版本为 v4，我们这个项目就用最新的版本来开发。所以，我们需要先下载 protoc-gen-micro v4 版本：


```shell

go install github.com/asim/go-micro/cmd/protoc-gen-micro/v4@latest


protoc -I $GOPATH/src  -I .  --micro_out=. --go_out=.  --go-grpc_out=.  hello.proto

# protoc -I $GOPATH/src  -I . --go_out=.  --go-grpc_out=.  hello.proto
```