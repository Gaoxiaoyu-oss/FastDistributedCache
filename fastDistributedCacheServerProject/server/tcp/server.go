package tcp

import (
	"FastMemCacheStore/fastDistributedCacheServerProject/server/cache/common"
	"FastMemCacheStore/fastDistributedCacheServerProject/server/cluster"
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

type Server struct {
	common.Cache
	cluster.Node //内嵌一个cluster.Node接口用来访问集群节点列表
}

func (s *Server) Listen() {
	l, e := net.Listen("tcp", s.Addr()+":12346")
	if e != nil {
		panic(e)
	}
	for {
		c, e := l.Accept()
		if e != nil {
			panic(e)
		}
		// 给每一个连接来的client都开启一个协程进行处理
		go s.process(c)
	}
}

// 改进为异步方式
func (s *Server) process(conn net.Conn) {
	// 在conn上套一层bufio.Reader，用于对客户端连接进行一个缓冲读取
	r := bufio.NewReader(conn)
	resultCh := make(chan chan *result, 5000)
	defer close(resultCh)

	go reply(conn, resultCh)

	for {
		//读取command中的op部分
		op, e := r.ReadByte()
		if e != nil {
			if e != io.EOF {
				log.Println("close connection due to error:", e)
			}
			return
		}
		if op == 'S' {
			s.set(resultCh, r)
		} else if op == 'G' {
			s.get(resultCh, r)
		} else if op == 'D' {
			s.del(resultCh, r)
		} else {
			log.Println("close connection due to invalid operation", op)
			return
		}
		if e != nil {
			log.Println("close connection due to error:", e)
			return
		}
	}
}

func reply(conn net.Conn, resultCh chan chan *result) {
	defer conn.Close()
	for {
		// 这里可能会阻塞等待，直到resultCh中有元素，或者resultCh已经被关闭
		c, open := <-resultCh
		if !open {
			return
		}
		// 等待缓存操作的结果
		r := <-c
		// 将结果发送给客户端的tcp连接conn
		e := sendResponse(r.v, r.e, conn)
		if e != nil {
			log.Println("close connection due to error:", e)
			return
		}
	}
}

// 根据参数将服务端的error或value 写入客户端连接
func sendResponse(value []byte, err error, conn net.Conn) error {
	if err != nil {
		errString := err.Error()
		tmp := fmt.Sprintf("-%d ", len(errString)) + errString
		_, e := conn.Write([]byte(tmp))
		return e
	}
	vlen := fmt.Sprintf("%d ", len(value))
	_, e := conn.Write(append([]byte(vlen), value...))
	return e
}

// 以空格为分割符并将之转化为一个整型
func readLen(r *bufio.Reader) (int, error) {
	tmp, e := r.ReadString(' ')
	if e != nil {
		return 0, e
	}
	l, e := strconv.Atoi(strings.TrimSpace(tmp))
	if e != nil {
		return 0, e
	}
	return l, nil
}

// 解析客户端的command，从中获取key和value
func (s *Server) readKey(r *bufio.Reader) (string, error) {
	klen, e := readLen(r)
	if e != nil {
		return "", e
	}
	k := make([]byte, klen)
	_, e = io.ReadFull(r, k)
	if e != nil {
		return "", e
	}
	key := string(k)
	// 查找一致性哈希，判断是否应该由本节点处理该key的缓存请求操作
	addr, ok := s.ShouldProcess(key)
	if !ok {
		return "", errors.New("redirect " + addr)
	}
	return key, nil
}

func (s *Server) readKeyAndValue(r *bufio.Reader) (string, []byte, error) {
	klen, e := readLen(r)
	if e != nil {
		return "", nil, e
	}
	vlen, e := readLen(r)
	if e != nil {
		return "", nil, e
	}
	k := make([]byte, klen)
	_, e = io.ReadFull(r, k)
	if e != nil {
		return "", nil, e
	}

	key := string(k)
	// 查找一致性哈希，判断是否应该由本节点处理该key的缓存请求操作
	addr, ok := s.ShouldProcess(key)
	if !ok {
		return "", nil, errors.New("redirect " + addr)
	}
	//如果应该由本节点进行处理:
	v := make([]byte, vlen)
	_, e = io.ReadFull(r, v)
	if e != nil {
		return "", nil, e
	}
	return key, v, nil

}

func (s *Server) get(ch chan chan *result, r *bufio.Reader) {
	c := make(chan *result)
	// TODO-假如有多个get请求，那么哪个请求先来，哪个请求的这个get方法创建的c就进入ch
	ch <- c
	k, e := s.readKey(r)
	if e != nil {
		c <- &result{nil, e}
		return
	}
	go func() {
		v, e := s.Get(k)
		c <- &result{v, e}
	}()
}

func (s *Server) set(ch chan chan *result, r *bufio.Reader) {
	c := make(chan *result)
	// TODO-假如有多个set请求，那么哪个请求先来，哪个请求的这个set方法创建的c就进入ch
	ch <- c
	k, v, e := s.readKeyAndValue(r)
	if e != nil {
		c <- &result{nil, e}
		return
	}
	go func() {
		c <- &result{nil, s.Set(k, v)}
	}()
}

func (s *Server) del(ch chan chan *result, r *bufio.Reader) {
	c := make(chan *result)
	// TODO-假如有多个del请求，那么哪个请求先来，哪个请求的这个del方法创建的c就进入ch
	ch <- c
	k, e := s.readKey(r)
	if e != nil {
		c <- &result{nil, e}
		return
	}
	go func() {
		c <- &result{nil, s.Del(k)}
	}()
}

func New(c common.Cache, n cluster.Node) *Server {
	return &Server{c, n}
}
