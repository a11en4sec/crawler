```shell

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

```shell

protoc -I $GOPATH/src  -I . --go_out=.  --go-grpc_out=.  hello.proto
```