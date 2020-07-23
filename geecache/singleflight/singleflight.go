package singleflight

import "sync"

// call 代表正在进行中 货已经结束的请求 使用sync.WaitGroup 锁避免重入
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group 是 singleflight 的主数据结构，管理不同key 的请求（call）
type Group struct {
	mu sync.Mutex // 保护 Group 的成员变量 不被并发读写而加上的锁
	m  map[string]*call
}

// Do 方法 第一个参数是key 第二个参数是一个函数 fn
// Do 的作用是，针对相同的key，无论Do被调用多少次，函数fn都只会被调用一次，等待fn调用结束了，翻翻返回值或错误

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock() // 加锁
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	/* 若请求正在处理中  这个判断是为了 多次请求不重复添加*/
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()       //  解锁
		c.wg.Wait()         // 如果请求正在进行中 则等待
		return c.val, c.err // 请求结果， 返回结果
	}

	/* 若第一次发起请求 添加到处理列表 */
	c := new(call)
	c.wg.Add(1)   // 发起请求前加锁
	g.m[key] = c  // 添加到 g.m 中， 表明 key 已经有对应的请求在处理
	g.mu.Unlock() // 解锁

	c.val, c.err = fn() // 调用fn，发起请求
	c.wg.Done()         // 请求结束

	g.mu.Lock()
	delete(g.m, key) // 更新 g.m
	g.mu.Unlock()

	return c.val, c.err // 返回结果
}
