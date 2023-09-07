package common

type Cache interface {
	Set(string, []byte) error
	Get(string) ([]byte, error)
	Del(string) error
	GetStat() Stat
	NewScanner() Scanner //新增方法，返回一个Scanner接口
}
