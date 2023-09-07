package inMemoryCache

import (
	"FastMemCacheStore/fastDistributedCacheServerProject/server/cache/common"
	"sync"
	"time"
)

type value struct {
	v       []byte    //实际保存的值的数据
	created time.Time //该值的创建时间，也就是上一次Set的时间
}

type inMemoryCache struct {
	c     map[string]value
	mutex sync.RWMutex
	common.Stat
	ttl time.Duration // 缓存生存时间
}
