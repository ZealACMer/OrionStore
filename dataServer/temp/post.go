package temp

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// tempInfo 记录临时对象的uuid、名字和大小
type tempInfo struct {
	Uuid string
	Name string
	Size int64
}

// post 用于处理HTTP的POST请求
func post(w http.ResponseWriter, r *http.Request) {
	//生成一个随机的uuid
	output, _ := exec.Command("uuidgen").Output()
	uuid := strings.TrimSuffix(string(output), "\n")
	//从请求的URL获取对象的名字，也就是散列值
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	//从Size头部读取对象的大小
	size, e := strconv.ParseInt(r.Header.Get("size"), 0, 64)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//拼成一个tempInfo结构体
	t := tempInfo{uuid, name, size}
	//调用tempInfo的writeToFile方法将结构体的内容写入磁盘上的文件
	e = t.writeToFile()
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//保存的临时对象的内容
	os.Create(os.Getenv("STORAGE_ROOT") + "/temp/" + t.Uuid + ".dat")
	//将该uuid通过HTTP响应返回给接口服务
	w.Write([]byte(uuid))
}

func (t *tempInfo) writeToFile() error {
	f, e := os.Create(os.Getenv("STORAGE_ROOT") + "/temp/" + t.Uuid)
	if e != nil {
		return e
	}
	defer f.Close()
	//将tempInfo的内容经过JSON编码后，写入该<uuid>文件
	//注意，这个文件是用于保存临时对象信息的，跟用于保存对象内容的<uuid>.dat是不同的两个文件
	b, _ := json.Marshal(t)
	f.Write(b)
	return nil
}
