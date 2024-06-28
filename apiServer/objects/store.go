package objects

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"orionStore/apiServer/locate"
	"orionStore/src/lib/utils"
)

func storeObject(r io.Reader, hash string, size int64) (int, error) {
	//定位对象的散列值，如已存在，则跳过后续上传操作直接返回200 OK
	if locate.Exist(url.PathEscape(hash)) {
		return http.StatusOK, nil
	}
	//生成对象的写入流stream，stream类型为*objectstream.PutStream
	stream, e := putStream(url.PathEscape(hash), size)
	if e != nil {
		return http.StatusInternalServerError, e
	}
	reader := io.TeeReader(r, stream)
	d := utils.CalculateHash(reader)
	//将计算的散列值与对象的散列值hash进行比较，如果不一致，则调用stream.Commit(false)删除临时对象，并返回400 Bad Request
	if d != hash {
		stream.Commit(false)
		return http.StatusBadRequest, fmt.Errorf("object hash mismatch, calculated=%s, requested=%s", d, hash)
	}
	//如果一致，则调用stream.Commit(true)将临时对象转正并返回200 OK
	stream.Commit(true)
	return http.StatusOK, nil
}
