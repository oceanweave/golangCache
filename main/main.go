package main

import (
	"Cache/geecache"
	"fmt"
	"log"
	"net/http"
)

// 首先 用一个map 模拟耗时的数据库
var db = map[string]string {
	"Tom": "630",
	"Jack": "345",
	"Sam": "562",
}

func main() {
	geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok { // 取到数据后返回
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exit", key)
		}))
	addr := "localhost:9999"
	peers := geecache.NewHTTPPool(addr)
	log.Println("Cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers)) // 第二个参数  是 实现ServeHTTP方法的接口
}
