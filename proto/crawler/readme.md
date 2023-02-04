### 1 根据proto文件生成代码

```shell
protoc -I $GOPATH/src  -I .  --micro_out=. --go_out=.  --go-grpc_out=.  --grpc-gateway_out=logtostderr=true,allow_delete_body=true,register_func_suffix=Gw:. crawler.proto
```
> body 字段报错,暂时注释

### 2 让 Master 实现 micro 生成的 CrawlerMasterHandler 接口
```go
type CrawlerMasterHandler interface {
  AddResource(context.Context, *ResourceSpec, *NodeSpec) error
  DeleteResource(context.Context, *ResourceSpec, *empty.Empty) error
}
```