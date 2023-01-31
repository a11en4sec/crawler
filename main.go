package main

import (
	_ "net/http/pprof"

	"github.com/a11en4sec/crawler/cmd"
)

func main() {
	cmd.Execute()
}

// ./main master
// ./main master --id=2 --http=:8082 --grpc=:9092 --pprof=:9982
// ./main master --id=3 --http=:8083 --grpc=:9093 --pprof=:9983
