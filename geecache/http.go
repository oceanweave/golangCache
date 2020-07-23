package geecache

import (
	"Cache/geecache/consistenthash"
	pb "Cache/geecache/geecachepb"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

/*
Go语言动手写Web框架 - Gee第一天 http.Handler
https://geektutu.com/post/gee-day1.html
http 库
https://golang.org/pkg/net/http/
*/
const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self     string //  用来记录自己的地址 包括主机名 IP 和端口
	basePath string // 节点间通讯地址的前缀 默认为上面的
	/* 添加节点选择功能 */
	mu          sync.Mutex             // 保护 peers 和 httpGetters
	peers       *consistenthash.Map    // 一致性哈希算法的Map 根据 key 选择节点
	httpGetters map[string]*httpGetter // keyed by e.g. "http://10.0.0.2:8008"
	// 映射远程节点对应的 httpGetter 每个远程节点对应一个 httpGetter 因为 httpGetter 与远程节点的地址 baseURL 有关
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// 最为核心的ServeHTTP 方法
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) { // 首先判断路径的前缀是否是 basePath
		panic("HTTPPool serving unexpected path:" + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path) // 方法 + url
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]
	group := GetGroup(groupName) // 通过 groupName 获得group实例
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}
	view, err := group.Get(key) // 获取缓存数据
	// proto新增
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	//w.Write(view.ByteSlice())
	w.Write(body) // proto 新增
}

/* 上面是服务端 */

/* 下面实现客户端 */
type httpGetter struct {
	baseURL string // 表示要访问的远程节点的地址 如http://example.com/_geecache/
}

// 获取返回值，并转化为[]bytes 类型
//func (h *httpGetter) Get(group string, key string) ([]byte, error) {
func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()))
	// fmt.Println("u", u)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returnes: %v", res.Status) // 没法访问到 返回错误
	}

	bytes, err := ioutil.ReadAll(res.Body)
	err = proto.Unmarshal(bytes, out)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}
	return nil
}

var _ PeerGetter = (*httpGetter)(nil) // 为了用来确保 htppGetter 实现了 PeerGetter接口

/* 实现 PeerPicker 接口 */
// Set 方法 实例化了一致性哈希算法，并添加了传入的节点
func (p *HTTPPool) Set(peers ...string) {
	// fmt.Println(peers)  // [http://localhost:8001 http://localhost:8002 http://localhost:8003]
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
		//fmt.Println("测试",peer) // 测试 http://localhost:8001
		//fmt.Println(*p.httpGetters[peer]) // {http://localhost:8001/_geecache/}
	}

}

//  PickPeer picks a peer according to key
// PickPeer 包装了一致性哈希算法的Get()方法，根据具体的key，选择节点，返回节点对应的HTTP客户端
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// fmt.Println(key, "对应的peer为", p.peers.Get(key))
	if peer := p.peers.Get(key); peer != "" && peer != p.self { // 判断 peer 不能为空 且不能为自己
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	// fmt.Println("没有找到合适的Peer（peer为自己 本地取数据就可以）") // 一个就是没有找到peer 另外就是peer是自己
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil) // 验证  HTTPPool 是否实现了PeerPicker 接口
