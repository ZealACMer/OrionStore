package versions

import (
	"encoding/json"
	"log"
	"net/http"
	"orionStore/src/lib/es"
	"strings"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	if m != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed) //405 Method Not Allowed
		return
	}
	from := 0
	size := 1000
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	for {
		//es.SearchAllVersions函数返回一个元数据的数组
		metas, e := es.SearchAllVersions(name, from, size)
		if e != nil {
			log.Println(e)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		//遍历该元数据数组，将元数据写入HTTP响应的正文
		for i := range metas {
			b, _ := json.Marshal(metas[i])
			w.Write(b)
			w.Write([]byte("\n"))
		}
		//如果返回的数组长度不等于size，说明元数据服务中没有更多的数据了，直接返回
		if len(metas) != size {
			return
		}
		//否则，就把from的值增加1000进行下一次迭代
		from += size
	}
}
