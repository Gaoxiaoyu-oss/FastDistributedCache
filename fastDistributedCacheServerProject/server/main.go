package main

import (
	"FastMemCacheStore/fastDistributedCacheServerProject/server/cache"
	"FastMemCacheStore/fastDistributedCacheServerProject/server/cluster"
	"FastMemCacheStore/fastDistributedCacheServerProject/server/http"
	"FastMemCacheStore/fastDistributedCacheServerProject/server/tcp"
	"flag"
	"log"
)

func main() {
	typ := flag.String("type", "inmemory", "cache type")     //本节点要开启的缓存服务类型--内存map/RocksDB
	ttl := flag.Int("ttl", 30, "cache time to live")         //缓存生存时间,单位为秒
	node := flag.String("node", "127.0.0.1", "node address") //本节点地址
	clus := flag.String("cluster", "", "cluster address")    //本节点要加入的集群中的任意一个节点的地址
	flag.Parse()
	log.Println("type is", *typ)
	log.Println("ttl is", *ttl)
	log.Println("node is", *node)
	log.Println("cluster is", *clus)
	c := cache.New(*typ, *ttl)
	// 给本节点创建集群所需的配置并开启相关服务
	n, e := cluster.New(*node, *clus)
	if e != nil {
		panic(e)
	}
	go tcp.New(c, n).Listen()
	http.New(c, n).Listen()
}
