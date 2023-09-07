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
	"errors"
	"regexp"
	"strconv"
	"time"
	"unsafe"
)

const BATCH_SIZE = 100

// 实现cache.Cache的Get方法
func (c *rocksdbCache) Get(key string) ([]byte, error) {
	// 用C.CString生成C语言的char*类型k
	k := C.CString(key)
	// 在函数退出时用free释放k的内存
	defer C.free(unsafe.Pointer(k))
	var length C.size_t
	v := C.rocksdb_get(c.db, c.ro, k, C.size_t(len(key)), &length, &c.e)

	if c.e != nil {
		return nil, errors.New(C.GoString(c.e))
	}
	defer C.free(unsafe.Pointer(v))
	return C.GoBytes(unsafe.Pointer(v), C.int(length)), nil
}

// 实现cache.Cache的Del方法，将指定key删除
func (c *rocksdbCache) Del(key string) error {
	k := C.CString(key)
	defer C.free(unsafe.Pointer(k))
	C.rocksdb_delete(c.db, c.wo, k, C.size_t(len(key)), &c.e)
	if c.e != nil {
		return errors.New(C.GoString(c.e))
	}
	return nil
}

//实现cache.Cache的GetStat方法
/*
RocksDB和Go语言的map不一样，它的内容不会在重启后丢失，也就是Stat不能从头开始计数，我们通过rocksdb_property_value获取rocksdb.aggregated-table-properties属性

这里获取的属性来自RocksDB的SST表文件，它和真实的数据相比有一定的滞后性：
RocksDB为了效率，使用了WAL技术，写入操作会不做处理地将数据尽快写入日志里，后台有线程慢慢处理日志内容并将其插入SST里面
*/
func (c *rocksdbCache) GetStat() common.Stat {
	k := C.CString("rocksdb.aggregated-table-properties")
	defer C.free(unsafe.Pointer(k))
	v := C.rocksdb_property_value(c.db, k)
	defer C.free(unsafe.Pointer(v))
	p := C.GoString(v)
	r := regexp.MustCompile(`([^;]+)=([^;]+);`)
	s := common.Stat{}
	for _, submatches := range r.FindAllStringSubmatch(p, -1) {
		if submatches[1] == " # entries" {
			s.Count, _ = strconv.ParseInt(submatches[2], 10, 64)
		} else if submatches[1] == " raw key size" {
			s.KeySize, _ = strconv.ParseInt(submatches[2], 10, 64)
		} else if submatches[1] == " raw value size" {
			s.ValueSize, _ = strconv.ParseInt(submatches[2], 10, 64)
		}
	}
	return s
}

func (cache *rocksdbCache) Set(key string, value []byte) error {
	cache.ch <- &common.Pair{key, value}
	return nil
}

func flush_batch(db *C.rocksdb_t, b *C.rocksdb_writebatch_t, o *C.rocksdb_writeoptions_t) {
	var e *C.char
	C.rocksdb_write(db, o, b, &e)
	if e != nil {
		panic(C.GoString(e))
	}
	C.rocksdb_writebatch_clear(b)
}

func write_func(db *C.rocksdb_t, c chan *common.Pair, o *C.rocksdb_writeoptions_t) {
	// 当前批次需要写入的键值对数量计数器
	count := 0
	// 1s后触发的计时器
	t := time.NewTimer(time.Second)
	// 用于批量写入RocksDB的结构体指针
	b := C.rocksdb_writebatch_create()

	for {
		select {
		case p := <-c:
			count++
			key := C.CString(p.K)
			value := C.CBytes(p.V)
			// 将键值对放入结构体指针b中pair
			C.rocksdb_writebatch_put(b, key, C.size_t(len(p.K)), (*C.char)(value), C.size_t(len(p.V)))
			C.free(unsafe.Pointer(key))
			C.free(value)
			// 如果此时批量数据达到了100个，则调用flush_batch将该批次写入RocksDB
			if count == BATCH_SIZE {
				flush_batch(db, b, o)
				count = 0
			}
			// 然后重置计数器和计时器
			if !t.Stop() {
				<-t.C
			}
			t.Reset(time.Second)

		//计时器触发时间为1s，一旦1s内没有后续写入，我们就会去批量写入，因此即使宕机，最多也只会丢失1s内的数据，且不超过100个
		case <-t.C:
			// 只要计数器不为0,则说明b中有需要写入的数据，此时也调用flush_batch进行批次写入并重置计数器和计时器
			if count != 0 {
				flush_batch(db, b, o)
				count = 0
			}
			t.Reset(time.Second)

		}
	}
}

func (c *rocksdbCache) NewScanner() common.Scanner {
	return &rocksdbScanner{C.rocksdb_create_iterator(c.db, c.ro), false}
}
