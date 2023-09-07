# FastDistributedCache

存储：
    内存map实现的缓存服务---支持缓存生存时间、支持tcp异步操作优化读取性能、支持集群(一致性hash、gossip)、支持节点负载再平衡
    RocksDB实现的缓存服务---支持数据持久化、支持缓存生存时间、支持批量写入、支持tcp异步操作优化读取性能、支持集群(一致性hash、gossip)、支持节点负载再平衡

    网络:
    tcp:主要用来向缓存服务发送操作请求(Set/Get/Del/GetStat)   支持异步操作，支持pipeline
    http：主要用于向系统发送控制信息(getStat/rebalance/)
