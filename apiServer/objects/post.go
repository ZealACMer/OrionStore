package objects

import (
	"log"
	"net/http"
	"net/url"
	"orionStore/apiServer/heartbeat"
	"orionStore/apiServer/locate"
	"orionStore/src/lib/es"
	"orionStore/src/lib/rs"
	"orionStore/src/lib/utils"
	"strconv"
	"strings"
)

// post 处理POST请求
func post(w http.ResponseWriter, r *http.Request) {
	//获取对象名字
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	//获取对象大小
	size, e := strconv.ParseInt(r.Header.Get("size"), 0, 64)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	//获取对象的散列值
	hash := utils.GetHashFromHeader(r.Header)
	if hash == "" {
		log.Println("missing object hash in digest header")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//通过散列值进行定位，判断散列值存在
	if locate.Exist(url.PathEscape(hash)) {
		//元数据服务添加新版本
		e = es.AddVersion(name, hash, size)
		if e != nil {
			log.Println(e)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		return
	}
	//散列值如果不存在，随机选择6个数据节点
	ds := heartbeat.ChooseRandomDataServers(rs.ALL_SHARDS, nil)
	if len(ds) != rs.ALL_SHARDS {
		log.Println("cannot find enough dataServer")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	//生成数据流stream，stream类型为指向RSResumablePutStream结构体的指针
	stream, e := rs.NewRSResumablePutStream(ds, name, url.PathEscape(hash), size)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//返回字符串token
	w.Header().Set("location", "/temp/"+url.PathEscape(stream.ToToken()))
	w.WriteHeader(http.StatusCreated)
}
