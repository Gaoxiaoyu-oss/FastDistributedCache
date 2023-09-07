package cache

// #include "rocksdb/c.h"
// #cgo CFLAGS: -I${SRCDIR}/../../../rocksdb/include
// #cgo LDFLAGS: -L${SRCDIR}/../../../rocksdb -lrocksdb -lz -lpthread -lsnappy -lstdc++ -lm -O3
/*
#include<stdlib.h>
*/
import "C"
import (
	"FastMemCacheStore/fastDistributedCacheServerProject/server/cache/common"
	"runtime"
)

func NewRocksdbCache(ttl int) *rocksdbCache {
	options := C.rocksdb_options_create()
	C.rocksdb_options_increase_parallelism(options, C.int(runtime.NumCPU()))
	C.rocksdb_options_set_create_if_missing(options, 1)
	var e *C.char
	// 将ttl参数传入，RocksDB自己管理缓存生存时间
	db := C.rocksdb_open_with_ttl(options, C.CString("/mnt/rocksdb"), C.int(ttl), &e)
	if e != nil {
		panic(C.GoString(e))
	}
	C.rocksdb_options_destroy(options)
	c := make(chan *common.Pair, 5000)
	wo := C.rocksdb_writeoptions_create()
	// 开启goroutine时刻准备批量写入
	go write_func(db, c, wo)
	return &rocksdbCache{db, C.rocksdb_readoptions_create(), wo, e, c}
}
