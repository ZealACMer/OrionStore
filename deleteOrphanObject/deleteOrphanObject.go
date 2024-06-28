package main

import (
	"github.com/robfig/cron/v3"
	"log"
	"net/http"
	"orionStore/src/lib/es"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	c := cron.New(cron.WithSeconds())                 //创建一个新的cron实例
	_, e := c.AddFunc("0 5 4 * * *", delOrphanObject) //每天的凌晨4点5分执行定时任务
	if e != nil {
		log.Println(e)
		return
	}
	c.Start() //开始定时任务

	select {}
}

// delOrphanObject 删除没有元数据引用的对象数据，需要在每个数据节点上定期运行
func delOrphanObject() {
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/*")

	for i := range files {
		hash := strings.Split(filepath.Base(files[i]), ".")[0]
		//调用es.HasHash函数检查元数据服务中是否存在该散列值
		hashInMetadata, e := es.HasHash(hash)
		if e != nil {
			log.Println(e)
			return
		}
		if !hashInMetadata {
			del(hash)
		}
	}
}

// del 调用数据服务的DELETE对象接口进行散列值的删除
func del(hash string) {
	log.Println("delete", hash)
	url := "http://" + os.Getenv("LISTEN_ADDRESS") + "/objects/" + hash
	request, _ := http.NewRequest("DELETE", url, nil)
	client := http.Client{}
	client.Do(request)
}
