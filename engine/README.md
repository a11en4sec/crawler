## 引擎调用过程v1
```

                Seeds
                   |
                   |       req           ParseFunc(req) HandleResult
requestCh----> reqQueue -----> workerCh ----------> out-----------> result:
^                                                                       - result.item ==> 存储
|                                                                       - result.req |
|---------------<----------------------<---------------------<-----------------------|

```

## 引擎调用过程v2
```

```

## 数据结构





