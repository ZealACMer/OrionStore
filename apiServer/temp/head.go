package temp

import (
	"fmt"
	"log"
	"net/http"
	"orionStore/src/lib/rs"
	"strings"
)

func head(w http.ResponseWriter, r *http.Request) {
	token := strings.Split(r.URL.EscapedPath(), "/")[2]
	//根据token恢复stream
	stream, e := rs.NewRSResumablePutStreamFromToken(token)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	//获取当前对象大小
	current := stream.CurrentSize()
	if current == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("content-length", fmt.Sprintf("%d", current))
}
