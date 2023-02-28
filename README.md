# crawler整体架构
## 1 引擎迭代演进
### 1.1 引擎调用过程v1
```

                Seeds
                   |
                   |       req           ParseFunc(req) HandleResult
requestCh----> reqQueue -----> workerCh ----------> out-----------> result:
^                                                                       - result.item ==> 存储
|                                                                       - result.req |
|---------------<----------------------<---------------------<-----------------------|

```

### 1.2 引擎调用过程v2
```

                Seeds
                   |
                   |                req           (Pull)                  result=ParseFunc(body,r)        HandleResult()
requestCh--> reqQueue(普通队列）     ----->workerCh-------> CreateWorker() ------------------------> out通道--result.Items ==> DB 
^       |--> priReqQueue(优先级队列) ------>|     |-------> CreateWorker()             |
|                                               |--------> CreateWorker()         - result.Requestrts...
|                                               | ......                             |       
|                                                                                    |
|                                                                                    |
|---------------<----------------------<--------Push-------------<-------------------|

```

## 2 服务注册与资源管理

借助etcd和go-micro.

go-micro 提供的 registry 接口提供了诸多 API:
```go

type Registry interface {
  Init(...Option) error
  Options() Options
  Register(*Service, ...RegisterOption) error
  Deregister(*Service, ...DeregisterOption) error
  GetService(string, ...GetOption) ([]*Service, error)
  ListServices(...ListOption) ([]*Service, error)
  Watch(...WatchOption) (Watcher, error)
  String() string
}
```
## 3 容器部署
### 3.1 docker构建镜像
```
docker image build -t crawler:latest .
docker image ls | grep crawler
docker image inspect crawler:latest
```
### 3.2 启动容器
```
docker run -p 8081:8080 crawler:latest // 有网络隔离的限制
```

### 3.3 docker-compose
```
docker-compose up
```
> docker-compose up 将查找名称为 docker-compose.yml 的配置文件，如果你有自定义的配置文件，需要使用 -f 标志指定它。另外，使用 -d 标志可以在后台启动应用程序。

### 3.4 查看
```
docker images
docker ps
docker network ls
```

### 3.5 测试
```shell
curl -H "content-type: application/json" -d '{"id":"zjx","name": "douban_book_list"}' http://localhost:18082/crawler/resource
```

## 4 k8s中部署 Worker Deployment
### 4.1 打包镜像
```shell

docker image build -t crawler:local .
```

### 4.2 推送到k8s集群
```shell
k3d image import crawler:local -c demo

```

### 4.3 kubectl apply
```shell

kubectl apply -f crawl-worker.yaml
```

### 4.4 查看pods状态
```shell
kubectl get pods -o wide // 可以查看到IP
kubectl logs xxx
kubectl get deployment

```

### 4.5 删除
```shell
 kubectl delete deployment crawler-deployment --grace-period=0 --force
 kubectl delete deployment crawler-master-deployment --grace-period=0 --force
 
```

### 4.6 进入pod所在的容器
```shell
 kubectl exec -it crawler-deployment-577bfbdcd6-kf4fk  -n default -- /bin/sh
```

### 4.7 使用集群中的一个curl服务器进行测试
```shell
# 进入容器
kubectl exec -it mycurlpod -- sh

# 测试
curl -H "content-type: application/json" -d '{"name": "john"}' http://10.42.3.5:8080/greeter/hello

```

## 5 k8s中部署 Worker Service
查看service的IP
```shell
> kubectl get service
NAME           TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE                                                          ok  k3d-demo kube  20:27:20 
kubernetes     ClusterIP   10.43.0.1      <none>        443/TCP             11d
crawl-worker   ClusterIP   10.43.248.98   <none>        8080/TCP,9090/TCP   6s

```
在curlpod中请求service的IP,会转发都后端服务
```shell
 curl -H "content-type: application/json" -d '{"name": "john"}' http://10.43.248.98:8080/greeter/hello
```

```shell
kubectl get svc
kubectl delete svc/xxxx
```

```shell
// 删除configmap
kubectl delete -f configmap.yaml
```








