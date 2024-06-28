package main

import (
	"github.com/robfig/cron/v3"
	"log"
	"orionStore/apiServer/objects"
	"orionStore/src/lib/es"
	"orionStore/src/lib/utils"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	c := cron.New(cron.WithSeconds())                     //创建一个新的cron实例
	_, e := c.AddFunc("0 10 4 * * *", startObjectScanner) //每天的凌晨4点10分执行定时任务
	if e != nil {
		log.Println(e)
		return
	}
	c.Start() //开始定时任务

	select {}
}

// startObjectScanner 需要在数据节点上定期运行，用于检查数据
func startObjectScanner() {
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/*")

	for i := range files {
		hash := strings.Split(filepath.Base(files[i]), ".")[0]
		verify(hash)
	}
}

func verify(hash string) {
	log.Println("verify", hash)
	size, e := es.SearchHashSize(hash)
	if e != nil {
		log.Println(e)
		return
	}
	stream, e := objects.GetStream(hash, size)
	if e != nil {
		log.Println(e)
		return
	}
	d := utils.CalculateHash(stream)
	if d != hash {
		log.Printf("object hash mismatch, calculated=%s, requested=%s", d, hash)
	}
	stream.Close()
}
