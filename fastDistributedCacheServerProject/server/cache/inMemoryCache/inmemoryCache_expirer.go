package inMemoryCache

import "time"

func (c *inMemoryCache) expirer() {
	for {
		// 每次睡眠c.ttl秒时间
		time.Sleep(c.ttl)
		c.mutex.RLock()
		// 遍历全部map中的键值对
		for k, v := range c.c {
			c.mutex.RUnlock()
			// 如果当前遍历的键值对的创建时间加上缓存生存时间仍然小于当前时间，那么就说明其已经过期
			if v.created.Add(c.ttl).Before(time.Now()) {
				// 从map中删除该键值对
				c.Del(k)
			}
			c.mutex.RLock()
		}
		c.mutex.RUnlock()
	}
}
