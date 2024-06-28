package rs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"orionStore/src/lib/objectstream"
	"orionStore/src/lib/utils"
)

type resumableToken struct {
	Name    string   //对象名字
	Size    int64    //对象大小
	Hash    string   //对象的散列值
	Servers []string //保存对象的6个分片的数据节点地址
	Uuids   []string //保存对象的6个分片的uuid
}

type RSResumablePutStream struct {
	*RSPutStream
	*resumableToken
}

// NewRSResumablePutStream 参数：dataServers(保存数据节点地址的数组)，name(对象的名字)，hash(对象的散列值), size(对象的大小)
func NewRSResumablePutStream(dataServers []string, name, hash string, size int64) (*RSResumablePutStream, error) {
	//调用NewRSPutStream，创建指向RSPutStream结构体的指针
	putStream, e := NewRSPutStream(dataServers, hash, size)
	if e != nil {
		return nil, e
	}
	uuids := make([]string, ALL_SHARDS)
	for i := range uuids {
		//获取6个分片的uuid
		uuids[i] = putStream.writers[i].(*objectstream.TempPutStream).Uuid
	}
	//创建resumableToken结构体
	token := &resumableToken{name, size, hash, dataServers, uuids}
	return &RSResumablePutStream{putStream, token}, nil
}

func NewRSResumablePutStreamFromToken(token string) (*RSResumablePutStream, error) {
	b, e := base64.StdEncoding.DecodeString(token)
	if e != nil {
		return nil, e
	}

	var t resumableToken
	e = json.Unmarshal(b, &t)
	if e != nil {
		return nil, e
	}

	writers := make([]io.Writer, ALL_SHARDS)
	for i := range writers {
		writers[i] = &objectstream.TempPutStream{t.Servers[i], t.Uuids[i]}
	}
	enc := NewEncoder(writers)
	return &RSResumablePutStream{&RSPutStream{enc}, &t}, nil
}

func (s *RSResumablePutStream) ToToken() string {
	//转换为JSON格式
	b, _ := json.Marshal(s)
	//返回经过Base64编码后的字符串
	return base64.StdEncoding.EncodeToString(b)
}

// CurrentSize 以HEAD方法获取第一个临时分片对象的大小，乘以4作为size返回
func (s *RSResumablePutStream) CurrentSize() int64 {
	r, e := http.Head(fmt.Sprintf("http://%s/temp/%s", s.Servers[0], s.Uuids[0]))
	if e != nil {
		log.Println(e)
		return -1
	}
	if r.StatusCode != http.StatusOK {
		log.Println(r.StatusCode)
		return -1
	}
	size := utils.GetSizeFromHeader(r.Header) * DATA_SHARDS
	//以对象的大小为上界
	if size > s.Size {
		size = s.Size
	}
	return size
}
