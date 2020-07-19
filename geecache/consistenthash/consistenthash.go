package consistenthash

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
)

//采取依赖注入的方式 允许用于提花昵称自定义的 Hash 函数 默认为crc32.ChecksumIEEE
type Hash func(data []byte) uint32

// Map 是一致性哈希算法的主数据结构
type Map struct {
	hash Hash // Hash 函数 hash
	replicas int // 虚拟节点倍数 replicas
	keys []int // 哈希环 keys
	hashMap map[int]string // 虚拟节点和真实节点的映射表 hashMap
	// 键是虚拟节点的哈希值 值是真实节点的名称
}

// New() 允许自定义虚拟节点倍数 和 Hash 函数
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas:replicas,
		hash:fn,
		hashMap:make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 实现添加真实节点/机器的 Add() 方法
/*
1. Add 允许传入 0 或 多个真实节点的名称
2. 对于每一个真实节点key， 对应创建 m.replicas 个虚拟节点
   虚拟节点的名称是strconv.Itoa(i)+key)， 即通过添加编号的方式区分不同虚拟节点
3. 使用 m.hash() 计算虚拟节点的哈希值，使用append(m.keys, hash) 添加到环上
4. 在hashMap 中增加虚拟节点 和 真实节点的映射关系
5. 最后一步，环上的哈希值排序
 */
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i)+key))) // 虚拟节点的哈希值
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys) // 排序
}
// 实现选择节点的 Get() 方法
/*
1. 计算 key 的哈希值
2. 顺时针找到第一个匹配的虚拟节点的下表 idx， 从 m.keys 中获取到对应的哈希值
   若 idx == len(m.keys) 说明应选择 m.keys[0]
   因为 m.keys 是一个环状结构，所以用取余数的方式来处理这种情况
3. 通过 hashMap 映射得到真实的节点
 */
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool { // 二分查找法 寻找满足函数的 最小索引
		return m.keys[i] >= hash
	})
	fmt.Println("key", key, "idx",idx)

	return m.hashMap[m.keys[idx%len(m.keys)]]
}

/*
https://studygolang.com/pkgdoc
Search函数采用二分法搜索找到[0, n)区间内最小的满足f(i)==true的值i。
也就是说，Search函数希望f在输入位于区间[0, n)的前面某部分（可以为空）时返回假，
而在输入位于剩余至结尾的部分（可以为空）时返回真；
Search函数会返回满足f(i)==true的最小值i。如果没有该值，函数会返回n。
注意，未找到时的返回值不是-1，这一点和strings.Index等函数不同。
Search函数只会用区间[0, n)内的值调用f。
 */