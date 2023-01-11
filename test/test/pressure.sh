#!/bin/bash

# ./pressure.sh -test.v
# 其中 -test.v是参数，输出详情
go test -c # -c会生成可执行文件 (文件夹.test)

PKG=$(basename $(pwd))  # 获取当前路径的最后一个名字，即为文件夹的名字
echo $PKG
while true ; do
        export GOMAXPROCS=$[ 1 + $[ RANDOM % 128 ]] # 随机的GOMAXPROCS
        ./$PKG.test $@ 2>&1   # $@代表可以加入参数   2>&1代表错误输出到控制台
done

