package inMemoryCache

import (
	"FastMemCacheStore/fastDistributedCacheServerProject/server/cache/common"
	"sync"
	"time"
)

func NewInMemoryCache(ttl int) *inMemoryCache {
	c := &inMemoryCache{make(map[string]value), sync.RWMutex{}, common.Stat{}, time.Duration(ttl) * time.Second}
	// 如果设置的缓存生存时间 > 0
	if ttl > 0 {
		go c.expirer()
	}
	return c
}
