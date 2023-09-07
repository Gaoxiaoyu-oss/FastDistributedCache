package cache

// #include "rocksdb/c.h"
// #cgo CFLAGS: -I${SRCDIR}/../../../rocksdb/include
// #cgo LDFLAGS: -L${SRCDIR}/../../../rocksdb -lrocksdb -lz -lpthread -lsnappy -lstdc++ -lm -O3
/*
#include<stdlib.h>
*/
import "C"
import "FastMemCacheStore/fastDistributedCacheServerProject/server/cache/common"

type rocksdbCache struct {
	db *C.rocksdb_t
	ro *C.rocksdb_readoptions_t
	wo *C.rocksdb_writeoptions_t
	e  *C.char
	ch chan *common.Pair //有一个接收者函数会从该channel读取键值对并实现批量写入
}
