package inMemoryCache

import (
	"FastMemCacheStore/fastDistributedCacheServerProject/server/cache/common"
	"time"
)

func (c *inMemoryCache) Set(k string, v []byte) error {
	c.mutex.Lock()
	c.mutex.Unlock()
	// 将k,v设置进map中
	c.c[k] = value{v, time.Now()}
	// 更新c.Stat
	c.AddValueForSize(k, v)
	return nil
}

func (c *inMemoryCache) Get(k string) ([]byte, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	// 从map中读取出k对应的value
	return c.c[k].v, nil
}

func (c *inMemoryCache) Del(k string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	v, exist := c.c[k]
	if exist {
		// 从map中删除该k对应的键值对
		delete(c.c, k)
		// 更新c.Stat
		c.DelValueForSize(k, v.v)
	}
	return nil
}

func (c *inMemoryCache) GetStat() common.Stat {
	return c.Stat
}

func (c *inMemoryCache) NewScanner() common.Scanner {
	pairCh := make(chan *common.Pair)
	closeCh := make(chan struct{})

	//运行一个新的goroutine执行匿名函数，用range 遍历 c.c，将遍历到的键值对发送到pairCh中
	go func() {
		defer close(pairCh)
		c.mutex.RLock()
		for k, v := range c.c {
			c.mutex.RUnlock()
			select {
			// 当closeCh可读,也就是closeCh里面有了元素 (对应inMemoryScanner.Close()方法，关闭s.closeCh)
			case <-closeCh:
				return
			// 当pairCh可写，也就是在另一个地方有对pairCh元素的接收 (对应inMemoryScanner.Scan方法从s.pairCh中取出元素)
			case pairCh <- &common.Pair{k, v.v}:

			}
			c.mutex.RLock()
		}
		c.mutex.RUnlock()
	}()

	return &inMemoryScanner{common.Pair{}, pairCh, closeCh}
}
