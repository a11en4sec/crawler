                          
                Seeds
                   |
                   |       req           ParseFunc(req) HandleResult
requestCh----> reqQueue -----> workerCh ----------> out-----------> result:
^                                                                       - result.item ==> 存储
|                                                                       - result.req |
|---------------<----------------------<---------------------<-----------------------|