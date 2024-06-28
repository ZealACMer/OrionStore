package objects

import (
	"net/http"
	"orionStore/dataServer/locate"
	"os"
	"path/filepath"
	"strings"
)

// del 删除操作
func del(w http.ResponseWriter, r *http.Request) {
	hash := strings.Split(r.URL.EscapedPath(), "/")[2]
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/" + hash + ".*")
	if len(files) != 1 {
		return
	}
	//在定位对象缓存中删除散列值
	locate.Del(hash)
	//将对象文件移动至垃圾目录
	os.Rename(files[0], os.Getenv("STORAGE_ROOT")+"/garbage/"+filepath.Base(files[0]))
}
