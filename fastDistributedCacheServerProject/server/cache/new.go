package cache

import (
	"FastMemCacheStore/fastDistributedCacheServerProject/server/cache/common"
	"FastMemCacheStore/fastDistributedCacheServerProject/server/cache/inMemoryCache"
	cache "FastMemCacheStore/fastDistributedCacheServerProject/server/cache/rocksdbCache"
	"log"
)

func New(typ string, ttl int) common.Cache {
	var c common.Cache
	if typ == "inmemory" {
		c = inMemoryCache.NewInMemoryCache(ttl)
	}
	if typ == "rocksdb" {
		c = cache.NewRocksdbCache(ttl)
	}
	if c == nil {
		panic("unknown cache type " + typ)
	}
	log.Println(typ, "ready to serve")
	return c
}
