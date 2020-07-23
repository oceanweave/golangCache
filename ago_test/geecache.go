package geecache

import (
	"fmt"
	"log"
	"sync"
)

// 此部分负责与外部交互 控制缓存存储和获取的主流程

/*
思考？ 若缓存不存在，则应从数据源（文件、数据库等）获取数据并添加到缓存中，那么是否应该支持多种数据源的配置呢？
答：
不应该
1. 数据源种类太多，没办法一一实现
2. 扩展性不好
解决方案：
应该由用户决定，那么交给用户好了
回调函数：
当缓存不存在时，调用这个函数，得到元数据
 */

// 定义接口 Getter  和 回调函数 Get
type Getter interface {
	Get(key string) ([]byte, error)
}

// 函数实现 Getter 接口
/*
定义一个函数类型 F，并且实现接口 A 的方法，然后在这个方法中调用自己。
这是 Go 语言中将其他函数（参数返回值定义与 F 一致）转换为接口 A 的常用技巧
 */
// 定义函数类型 GetterFunc 并实现 Getter 接口的Get方法
// GetterFunc 定义拉取方法
type GetterFunc func(key string) ([]byte, error)

// Get 进行拉取
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// 核心数据结构 Group
/*
一个Group 可以认为是一个缓存的命名空间
每个Group拥有一个唯一的名称 name
比如可以创建三个Group
缓存学生成绩的命名为scores，缓存学生信息的命名为 info， 缓存学生课程的命名为 courses
getter 为 缓存未命中是获取源数据的回调 callback
mainCache  就是一开始实现的 并发缓存
 */
type Group struct {
	name string
	getter Getter
	mainCache cache
	peers PeerPicker // 增加分布式
}

// 全局变量
var (
	mu sync.RWMutex
	groups = make(map[string]*Group)
)

// 实例化 Group 并将 group 存储在全局变量 groups 中
// 参数为 name group 名字 cacheBytes 缓存空间大小 getter 回调函数
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:name,
		getter:getter,
		mainCache:cache{cacheBytes:cacheBytes},
	}
	groups[name] = g
	return g
}

// 用来获取特定名称的 group  只用到了 读锁
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// 接下来是 GeeCache 最为核心的方法Get
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" { // 判断key是否合法
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok { // 在本地缓存中查找
		fmt.Println("查找本地缓存")
		log.Println("[GeeCache] hit")
		return v, nil
	}
	return g.load(key) // 没找到 调用load 方法
}

func (g *Group) load(key string) (value ByteView, err error) {
	// 更新分布式场景
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value,err = g.getFromPeer(peer, key); err!= nil {
				return value, nil
			}
			log.Println("[GeeCache] Failed to get from peer", err)
		}
	}
	return g.getLoacally(key) // 从本地节点获取
	// 分布式场景下回调用 getFromPeer 从其他节点获取
}

func (g *Group) getLoacally(key string) (ByteView, error) {
	fmt.Println("从本地节点取数据")
	bytes, err := g.getter.Get(key) // 调用用户回调函数 获取源数据
	if err != nil {
		fmt.Println(err)
		return ByteView{},err
	}
	value := ByteView{b: cloneBytes(bytes)} // 返回拷贝
	g.populateCache(key, value) // 并将源数据添加到缓存中
	return value, nil
}

// 将数据添加到缓存中
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// RegisterPeers 注册一个 PeerPicker 用于选择远程Peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}



func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}