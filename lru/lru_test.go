package lru

import (
	"fmt"
	"reflect"
	"testing"
)

type String string // 相当于Value

func (d String) Len() int { // 相当于实现了 Value 接口
	return len(d)
}

func TestGet(t *testing.T) {
	lru := New(int64(0),nil)
	lru.Add("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failes")
	}
}

func TestCache_RemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2,String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("RemoveOldest key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	} // 回调函数作用 将删除的 key 添加到keys 切片中
	lru := New(int64(10), callback)
	lru.Add("key1", String("123456"))
	fmt.Println(len(String("123456")))
	lru.Add("k2",String("k2"))
	lru.Add("k3",String("k3"))
	lru.Add("k4",String("k4"))
	expect := []string{"key1","k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}