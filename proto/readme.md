## 1 protoc执行报错
```shell
Import “google/api/annotations.proto” was not found or had errors.
```
解决：
```shell
将包
https://github.com/googleapis/googleapis
下的google/
拷贝到$GOPATH/src/下

```

## 2 protoc执行报错
```google/protobuf/descriptor.proto: File not found.
google/api/annotations.proto:20:1: Import "google/protobuf/descriptor.proto" was not found or had errors.
google/api/annotations.proto:28:8: "google.protobuf.MethodOptions" is not defined.
hello.proto:3:1: Import "google/api/annotations.proto" was not found or had errors.
```
解决:
```shell
将包
https://github.com/protocolbuffers/protobuf
下的src/google/protobuf
拷贝到$GOPATH/src/google下

```

## 3 
```shell
        module declares its path as: github.com/micro/go-micro
                but was required as: go-micro.dev/v4/api

```
解决：
```shell
 go get go-micro.dev/v4
```