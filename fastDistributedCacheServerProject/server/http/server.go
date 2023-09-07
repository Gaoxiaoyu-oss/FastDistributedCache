package http

import (
	"FastMemCacheStore/fastDistributedCacheServerProject/server/cache/common"
	"FastMemCacheStore/fastDistributedCacheServerProject/server/cluster"
	"net/http"
)

type Server struct {
	common.Cache
	cluster.Node
}

func (s *Server) Listen() {
	http.Handle("/cache/", s.cacheHandler())
	http.Handle("/status", s.statusHandler())   //http客户端通过该接口获取当前节点存储的键值对统计信息
	http.Handle("/cluster", s.clusterHandler()) //http客户端通过该接口获取当前节点所在集群的信息
	http.Handle("/rebalance", s.rebalanceHandler())
	http.ListenAndServe(s.Addr()+":12345", nil)
}

func New(c common.Cache, n cluster.Node) *Server {
	return &Server{c, n}
}
