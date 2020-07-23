package ago_test

import (
	"Cache/lru"
	"sync"
)
// 此部分负责并发控制
type cache struct {
	mu sync.Mutex
	lru *lru.Cache
	cacheBytes int64 // maxBytes 允许的最大内存
}
/*
实例化lru
封装 get 和 add 方法 并添加互斥锁mu
 */
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil { // 这种叫 延迟初始化  主要用于提高性能 减少程序内存的要求
		c.lru = lru.New(c.cacheBytes, nil) // 实例化lru
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool ){
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}