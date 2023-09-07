package common

// inMemoryCache和rocksdbCache都要实现该接口
type Scanner interface {
	Scan() bool    //如果返回true,表示后续还有未遍历的键值对,如果返回false则表示遍历结束
	Key() string   //访问当前键值对的key
	Value() []byte //访问当前键值对的value
	Close()        //结束遍历
}
