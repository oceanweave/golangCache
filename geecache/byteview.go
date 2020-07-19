package geecache

// 此部分负责缓存值的抽象和封装
// 只读数据结构 ByteView 用来表示缓存值
type ByteView struct {
	b []byte // 真实的缓存值 选择byte类型是为了支持任意的数据类型存储 如字符 图片等
}

func (v ByteView) Len() int {
	return len(v.b)
	// lru.Cache 实现中 要求被缓存对象 必须实现 Value接口 即Len() int 方法  返回其所占用内存的大小
}

// b 是只读的 因此使用ByteSlice（) 方法返回一个拷贝， 繁殖缓存值被外部程序修改
// 返回 数据的拷贝 []byte 切片 封装cloneBytes
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}
// 返回数据的拷贝
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func (v ByteView) String() string {
	return string(v.b)
}