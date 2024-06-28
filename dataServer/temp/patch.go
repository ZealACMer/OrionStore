package temp

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// patch 用于数据上传
func patch(w http.ResponseWriter, r *http.Request) {
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	tempinfo, e := readFromFile(uuid)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	datFile := infoFile + ".dat"
	f, e := os.OpenFile(datFile, os.O_WRONLY|os.O_APPEND, 0)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()
	_, e = io.Copy(f, r.Body)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//写完后调用f.Stat方法获取数据文件的信息info
	info, e := f.Stat()
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//用info.Size获取数据文件当前的大小
	actual := info.Size()
	//如果超出了tempInfo中记录的大小，我们就删除信息文件和数据文件，并返回500 Internal Server Error
	if actual > tempinfo.Size {
		os.Remove(datFile)
		os.Remove(infoFile)
		log.Println("actual size", actual, "exceeds", tempinfo.Size)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// readFromFile 读取uuid对应的文件内容，并将内容通过JSON解码成一个tempInfo结构体返回
func readFromFile(uuid string) (*tempInfo, error) {
	f, e := os.Open(os.Getenv("STORAGE_ROOT") + "/temp/" + uuid)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, _ := io.ReadAll(f)
	var info tempInfo
	json.Unmarshal(b, &info)
	return &info, nil
}
