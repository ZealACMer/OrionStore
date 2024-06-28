package temp

import (
	"compress/gzip"
	"io"
	"net/url"
	"orionStore/dataServer/locate"
	"orionStore/src/lib/utils"
	"os"
	"strconv"
	"strings"
)

func (t *tempInfo) hash() string {
	s := strings.Split(t.Name, ".")
	return s[0]
}

func (t *tempInfo) id() int {
	s := strings.Split(t.Name, ".")
	id, _ := strconv.Atoi(s[1])
	return id
}

func commitTempObject(datFile string, tempinfo *tempInfo) {
	f, _ := os.Open(datFile)
	defer f.Close()
	d := url.PathEscape(utils.CalculateHash(f))
	f.Seek(0, io.SeekStart)
	//tempinfo.Name为对象的散列值
	w, _ := os.Create(os.Getenv("STORAGE_ROOT") + "/objects/" + tempinfo.Name + "." + d)
	w2 := gzip.NewWriter(w)
	io.Copy(w2, f)
	w2.Close()
	//删除临时对象文件
	os.Remove(datFile)
	//添加对象定位缓存
	locate.Add(tempinfo.hash(), tempinfo.id())
}
