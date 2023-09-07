package cluster

import (
	"github.com/hashicorp/memberlist"
	"io/ioutil"
	"stathat.com/c/consistent"
	"time"
)

type Node interface {
	ShouldProcess(key string) (string, bool) //接收一个string类型的key，用来告诉节点该key是否应该由自己处理
	Members() []string                       //用于提供整个集群的节点列表
	Addr() string                            //用于获取本节点的地址
}

// node实现Node接口,其中Members方法已经由consistent.Consistent实现
type node struct {
	//内嵌，继承的目的
	*consistent.Consistent
	//记录本节点地址
	addr string
}

// 让外部调用者获取本节点地址
func (n *node) Addr() string {
	return n.addr
}

// 调用consistent.Consistent的Get方法来获取可以处理这个Key的节点地址, 第一个返回值：处理该key的节点地址，第二个返回值：如果本节点就可以处理，则为true，否则为false
func (n *node) ShouldProcess(key string) (string, bool) {
	addr, _ := n.Get(key)
	return addr, addr == n.addr
}

func New(addr, cluster string) (Node, error) {
	// 这里我们用局域网配置
	conf := memberlist.DefaultLANConfig()
	// 将节点名字设置为命令行参数中的本节点地址
	conf.Name = addr
	// 将gossip监听地址设置为命令行参数中的本节点地址
	conf.BindAddr = addr
	// 设置日志输出器
	conf.LogOutput = ioutil.Discard

	// 创建memberlist.Memberlist结构体指针l
	l, e := memberlist.Create(conf)
	if e != nil {
		return nil, e
	}
	// 如果该参数为空，将本机地址作为集群节点，也就是集群只有本机一个节点
	if cluster == "" {
		cluster = addr
	}
	clu := []string{cluster}
	// 加入到命令行参数指定的集群中某节点所在的集群
	_, e = l.Join(clu)
	if e != nil {
		return nil, e
	}
	// 创建consistent.Consistent结构体指针circle
	circle := consistent.New()
	// 设置每个节点的虚拟节点个数
	circle.NumberOfReplicas = 256

	go func() {
		// 每隔1s,将memberlist.Memberlist.Members方法提供的集群节点列表m更新到circle中
		for {
			m := l.Members()
			nodes := make([]string, len(m))
			for i, n := range m {
				nodes[i] = n.Name
			}
			circle.Set(nodes)
			time.Sleep(time.Second)
		}
	}()
	return &node{circle, addr}, nil
}
