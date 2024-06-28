package objects

import (
	"log"
	"net/http"
	"orionStore/src/lib/es"
	"strings"
)

// del 删除对象
func del(w http.ResponseWriter, r *http.Request) {
	//获取对象名称
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	//搜索该对象的最新版本
	version, e := es.SearchLatestVersion(name)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//调用es.PutMetadata插入一条新的元数据，四个参数分别为元数据的name、version、size和hash,这是一个删除标记
	e = es.PutMetadata(name, version.Version+1, 0, "")
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
