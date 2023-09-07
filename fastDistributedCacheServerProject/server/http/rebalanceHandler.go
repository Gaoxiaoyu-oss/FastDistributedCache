package http

import (
	"bytes"
	"net/http"
)

// 实现http.Handler接口
type rebalanceHandler struct {
	*Server
}

func (h *rebalanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	go h.rebalance()
}

func (h *rebalanceHandler) rebalance() {
	// 获取缓存遍历器
	s := h.NewScanner()
	defer s.Close()
	c := &http.Client{}
	// 遍历自身的每一个键值对
	for s.Scan() {
		k := s.Key()
		// 判断该key是否应由本节点进行处理
		n, ok := h.ShouldProcess(k)
		// 如果不应该由本节点处理
		if !ok {
			// 则我们通过http来访问新节点的cache接口，将该键值对插入该新节点
			r, _ := http.NewRequest(http.MethodPut, "http://"+n+":12345/cache/"+k, bytes.NewReader(s.Value()))
			c.Do(r)
			// 然后我们在本节点删除该键值对
			h.Del(k)
		}
	}
}

func (s *Server) rebalanceHandler() http.Handler {
	return &rebalanceHandler{s}
}
