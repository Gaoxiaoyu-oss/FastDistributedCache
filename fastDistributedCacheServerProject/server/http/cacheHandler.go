package http

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type cacheHandler struct {
	// 内嵌了Server，相当于继承了该结构体
	*Server
}

// 实现该方法意味着实现了http.Handler接口
func (h *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := strings.Split(r.URL.EscapedPath(), "/")[2]
	if len(key) == 0 {
		// 如果key为空，则返回400错误码
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	m := r.Method
	// 如果是Set，对应Put方法
	if m == http.MethodPut {
		// 请求消息的请求体为put的value值
		b, _ := ioutil.ReadAll(r.Body)
		if len(b) != 0 {
			e := h.Set(key, b)
			if e != nil {
				log.Println(e)
				// 如果写入出错，则返回500错误码
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		return
	}
	// 如果是Get,对应Get方法
	if m == http.MethodGet {
		b, e := h.Get(key)
		if e != nil {
			log.Println(e)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(b) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// 如果能找到，通过httpReponse返回value值
		w.Write(b)
		return
	}
	//如果是del方法，对应DELETE方法
	if m == http.MethodDelete {
		e := h.Del(key)
		if e != nil {
			log.Println(e)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// 如果都不是GET/PUT/DELETE方法，那么就回复不支持的请求类型
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *Server) cacheHandler() http.Handler {
	return &cacheHandler{s}
}
