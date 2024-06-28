package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func GetOffsetFromHeader(h http.Header) int64 {
	byteRange := h.Get("range")
	if len(byteRange) < 7 {
		return 0
	}
	if byteRange[:6] != "bytes=" {
		return 0
	}
	bytePos := strings.Split(byteRange[6:], "-")
	offset, _ := strconv.ParseInt(bytePos[0], 0, 64)
	return offset
}

func GetHashFromHeader(h http.Header) string {
	digest := h.Get("digest")
	//检查digest的形式是否为"SHA-256=<hash>"
	if len(digest) < 9 {
		return ""
	}
	//如果不是以"SHA-256="开头，则返回空字符串
	if digest[:8] != "SHA-256=" {
		return ""
	}
	//否则返回"SHA-256="开头后面的部分
	return digest[8:]
}

func GetSizeFromHeader(h http.Header) int64 {
	size, _ := strconv.ParseInt(h.Get("content-length"), 0, 64)
	return size
}

func CalculateHash(r io.Reader) string {
	h := sha256.New()
	//io.Copy从参数r中读取数据并写入h，h会对写入的数据计算其散列值，这个散列值可以通过h.Sum方法读取。
	io.Copy(h, r)
	//从h.Sum读取到的散列值是一个二进制的数据，还需要用base64.StdEncoding.EncodeToString函数进行Base64编码
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
