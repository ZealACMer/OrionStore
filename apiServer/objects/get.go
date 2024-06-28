package objects

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"orionStore/src/lib/es"
	"orionStore/src/lib/utils"
	"strconv"
	"strings"
)

// get 用来处理GET请求
func get(w http.ResponseWriter, r *http.Request) {
	//获取对象的名字
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	//获取查询参数version的值
	versionId := r.URL.Query()["version"]
	version := 0
	var e error
	if len(versionId) != 0 {
		version, e = strconv.Atoi(versionId[0])
		if e != nil {
			log.Println(e)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	//调用es.GetMetadata，参数分别为对象的名字及版本号，返回值为对象的元数据
	meta, e := es.GetMetadata(name, version)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//meta.Hash为对象的散列值，如果散列值为空字符串说明对象该版本是一个删除标记，返回404 Not Found
	if meta.Hash == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	hash := url.PathEscape(meta.Hash)
	//以散列值为对象名从数据服务层获取对象并输出
	stream, e := GetStream(hash, meta.Size)
	//如果出现错误，则打印log并返回404 Not Found
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	//从HTTP请求的Range头部获得客户端要求的偏移量offset
	offset := utils.GetOffsetFromHeader(r.Header)
	if offset != 0 {
		//如果offset不为0，需要调用stream的Seek方法跳到offset位置
		stream.Seek(offset, io.SeekCurrent)
		w.Header().Set("content-range", fmt.Sprintf("bytes %d-%d/%d", offset, meta.Size-1, meta.Size))
		w.WriteHeader(http.StatusPartialContent)
	}
	acceptGzip := false
	encoding := r.Header["Accept-Encoding"]
	for i := range encoding {
		if encoding[i] == "gzip" {
			acceptGzip = true
			break
		}
	}
	//以Gzip方式获取数据
	if acceptGzip {
		w.Header().Set("content-encoding", "gzip")
		w2 := gzip.NewWriter(w)
		io.Copy(w2, stream)
		w2.Close()
	} else {
		//用io.Copy将stream的内容写入HTTP响应的正文
		io.Copy(w, stream)
	}
	stream.Close()
}
