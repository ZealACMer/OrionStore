package temp

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"orionStore/apiServer/locate"
	"orionStore/src/lib/es"
	"orionStore/src/lib/rs"
	"orionStore/src/lib/utils"
	"strings"
)

func put(w http.ResponseWriter, r *http.Request) {
	token := strings.Split(r.URL.EscapedPath(), "/")[2]
	//创建RSResumablePutStream结构体并获得指向该结构体的指针stream
	stream, e := rs.NewRSResumablePutStreamFromToken(token)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	//获取对象的大小
	current := stream.CurrentSize()
	if current == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	//获取offset
	offset := utils.GetOffsetFromHeader(r.Header)
	if current != offset {
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}
	//以32000字节为长度读取HTTP请求的正文并写入stream
	bytes := make([]byte, rs.BLOCK_SIZE)
	for {
		n, e := io.ReadFull(r.Body, bytes)
		if e != nil && e != io.EOF && e != io.ErrUnexpectedEOF {
			log.Println(e)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		current += int64(n)
		//读取的数据大小 > 对象大小
		if current > stream.Size {
			stream.Commit(false)
			log.Println("resumable put exceed size")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		//读取的数据不满32000字节，且已读到的数据 != 对象大小
		if n != rs.BLOCK_SIZE && current != stream.Size {
			return
		}
		stream.Write(bytes[:n])
		if current == stream.Size {
			stream.Flush()
			getStream, e := rs.NewRSResumableGetStream(stream.Servers, stream.Uuids, stream.Size)
			hash := url.PathEscape(utils.CalculateHash(getStream))
			//散列值不一致
			if hash != stream.Hash {
				stream.Commit(false)
				log.Println("resumable put done but hash mismatch")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			if locate.Exist(url.PathEscape(hash)) {
				stream.Commit(false)
			} else {
				stream.Commit(true)
			}
			e = es.AddVersion(stream.Name, stream.Hash, stream.Size)
			if e != nil {
				log.Println(e)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}
}
