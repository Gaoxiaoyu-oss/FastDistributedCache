package inMemoryCache

import "FastMemCacheStore/fastDistributedCacheServerProject/server/cache/common"

// 实现Scanner接口
type inMemoryScanner struct {
	common.Pair
	pairCh  chan *common.Pair //用于接收pair结构体指针的channel
	closeCh chan struct{}     //用于终止遍历的channel
}

// 关闭closeCh，此时closeCh的接收端将从阻塞中唤醒
func (s *inMemoryScanner) Close() {
	close(s.closeCh)
}

// 从s.pairCh中读取一个pair结构体指针p和一个bool变量ok,当s.pairCh被关闭时，读到的p为nil,ok为false
func (s *inMemoryScanner) Scan() bool {
	p, ok := <-s.pairCh //从s.pairCh取出数据-pair
	if ok {
		s.K, s.V = p.K, p.V
	}
	return ok
}

func (s *inMemoryScanner) Key() string {
	return s.K
}

func (s *inMemoryScanner) Value() []byte {
	return s.V
}
