package objects

import (
	"compress/gzip"
	"io"
	"log"
	"os"
)

func sendFile(w io.Writer, file string) {
	f, e := os.Open(file)
	if e != nil {
		log.Println(e)
		return
	}
	defer f.Close()
	//创建一个指向gzip.Reader结构体的指针gzipStream，对f中的数据进行解压
	gzipStream, e := gzip.NewReader(f)
	if e != nil {
		log.Println(e)
		return
	}
	io.Copy(w, gzipStream)
	gzipStream.Close()
}
