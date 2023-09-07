package http

import (
	"encoding/json"
	"log"
	"net/http"
)

type statusHandler struct {
	// 内嵌了Server，相当于继承了该结构体
	*Server
}

func (h *statusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	//调用cache.Cache.GetStat方法获取Stat，并将其用json进行编码成字节切片
	b, e := json.Marshal(h.GetStat())
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func (s *Server) statusHandler() http.Handler {
	return &statusHandler{s}
}
