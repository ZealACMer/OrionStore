package locate

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handler 用于处理HTTP请求
func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	if m != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed) //405 Method Not Allowed
		return
	}
	//Locate函数返回拥有请求对象分片的数据服务节点的地址
	info := Locate(strings.Split(r.URL.EscapedPath(), "/")[2])
	if len(info) == 0 {
		w.WriteHeader(http.StatusNotFound) //404 Not Found
		return
	}
	b, _ := json.Marshal(info)
	//将数据服务节点的地址作为HTTP响应的正文输出
	w.Write(b)
}
