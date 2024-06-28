package objects

import (
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	//r的Method成员变量记录HTTP请求的方法
	m := r.Method
	if m == http.MethodGet {
		get(w, r)
		return
	}
	if m == http.MethodDelete {
		del(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}
