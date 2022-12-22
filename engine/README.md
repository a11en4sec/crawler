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

## 数据结构





