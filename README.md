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










