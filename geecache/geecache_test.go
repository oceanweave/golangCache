package geecache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	}) // 相当于 定义一个函数 转为GetterFunc 函数类型
	// Getter 是接口  GetterFunc 实现了此接口

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("call back failed")
	}
}

// 首先 用一个map 模拟耗时的数据库
var db = map[string]string {
	"Tom": "630",
	"Jack": "345",
	"Sam": "562",
}

func TestGet(t *testing.T)  {
	loadCounts := make(map[string]int, len(db))
	gee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error){
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1

				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exit", key)
		}))
	for k, v := range db {
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		}
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			// 使用 loadCounts 统计某个键调用回调函数的次数，
			// 如果次数大于1，则表示调用了多次回调函数，没有缓存。
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unkown should be empty, but %s get", view)
	}
}