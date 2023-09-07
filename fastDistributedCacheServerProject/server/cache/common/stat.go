package common

// 缓存系统的统计信息
type Stat struct {
	// 目前保存的键值对数量
	Count int64
	// Key目前所占用的字节数
	KeySize int64
	// Value目前所占用的字节数
	ValueSize int64
}

// 新加键值对时改变缓存的状态
func (s *Stat) AddValueForSize(k string, v []byte) {
	s.Count += 1
	s.KeySize += int64(len(k))
	s.ValueSize += int64(len(v))
}

// 删除键值对时改变缓存的状态
func (s *Stat) DelValueForSize(k string, v []byte) {
	s.Count -= 1
	s.KeySize -= int64(len(k))
	s.ValueSize -= int64(len(v))
}
