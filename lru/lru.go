package lru

import "container/list"
/*
test 简明教程 https://geektutu.com/post/quick-go-test.html
list 官方文档 https://golang.org/pkg/container/list/

 */
type Cache struct {
	maxBytes int64 // 允许的最大内存
	nbytes int64	// 当前已使用的内存
	ll  *list.List // 双向链表
	cache map[string]*list.Element
	OnEvicted func(key string, value Value) // 某条记录被移除是的回调函数 可为nil
}

type entry struct { // 双向链表节点的数据类型
	key string
	value Value
}

type Value interface { //
	Len() int // 返回值所占用的内存大小
}

// 实例化
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:maxBytes,
		ll: list.New(),
		cache:make(map[string]*list.Element),
		OnEvicted:onEvicted,
	}
}

// 查找功能
func (c *Cache) Get(key string) (value Value, ok bool) {
	if elem, ok := c.cache[key]; ok {
		c.ll.MoveToFront(elem) // 移到队首
		kv := elem.Value.(*entry)
		return kv.value, true
	}
	return
}

// 删除
func (c *Cache) RemoveOldest() {
	elem := c.ll.Back() // 取出队尾元素
	if elem != nil {
		c.ll.Remove(elem) // 将队尾元素从链表中删除
		kv := elem.Value.(*entry)
		delete(c.cache, kv.key)  // 删除映射
		c.nbytes -= (int64((len(kv.key))) + int64(kv.value.Len())) // 更新已使用的内存
		if c.OnEvicted != nil { // 回调函数
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// 新增/修改
func (c *Cache) Add(key string, value Value) {
	if elem, ok := c.cache[key]; ok { // 存在 更新
		c.ll.MoveToFront(elem) // 移到队首
		kv := elem.Value.(*entry)
		c.nbytes += (int64(value.Len()) - int64(kv.value.Len())) // 更新已使用内存 新增多少内存
		kv.value = value
	} else { // 新增
		elem := c.ll.PushFront(&entry{key,value})
		c.cache[key] = elem
		c.nbytes += (int64(len(key)) + int64(value.Len()))
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes { // 若超过最大值 则移除最少访问的节点
		// 感觉此处逻辑有错误  应该是先判断 在添加 与上面两端对调
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int { // 列出缓存的条目数  双向链表中的条目数
	return c.ll.Len()
}

