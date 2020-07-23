package main

import (
	"Cache/geecache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "245",
}

func createGroup() *geecache.Group {
	return geecache.NewGroup("scores", 2<<10, geecache.GetterFunc( // 取数据函数 回调函数
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] c key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// 启动缓存服务器
/*
1. 创建HTTPPool，添加节点信息，注册到个恶中
2. 启动HTTP服务（共3个端口 8001 8002 8003） 用户不感知
*/
func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	peers := geecache.NewHTTPPool(addr)
	peers.Set(addrs...)
	gee.RegisterPeers(peers)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// 启动一个 API 服务 端口9999， 与用户进行交互，用户感知
func startAPIServer(apiAddr string, gee *geecache.Group) {
	fmt.Println("启动Api-server")
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			//fmt.Println("--------------")
			//fmt.Println("开始处理")
			key := r.URL.Query().Get("key")
			//fmt.Println("解析到的key为", key)
			view, err := gee.Get(key) // 得到 key 对应的数据 比如 630 或 返回错误 kkk not exist
			fmt.Println(view)
			if err != nil {
				// fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError) // 无法返回 错误
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice()) // 不知道为什么 后面多个% 比如630 打印出 630%
			// fmt.Println(view.ByteSlice(), n)
		}))
	log.Println("fontend server is runing at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
	// fmt.Println("结束处理")
}

// 需要命令函传入 port 和 api 2个参数， 用来在指定端口启动 HTTP 服务
func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := createGroup()
	if api {
		//
		fmt.Println("启动成功，此时 port", port)
		go startAPIServer(apiAddr, gee)
	}
	startCacheServer(addrMap[port], []string(addrs), gee)
}
