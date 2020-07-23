package geecache

import (
	pb "Cache/geecache/geecachepb"
)

// 用于根据传入的key 选择相应节点 PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// 用于从对应group查找缓存值，PeerGetter 对应于上述流程中的 HTTP 客户端
type PeerGetter interface {
	//Get(group string, key string) ([]byte, error)
	Get(in *pb.Request, out *pb.Response) error
}
