package main

import (
	"github.com/robfig/cron/v3"
	"log"
	"orionStore/src/lib/es"
)

const MIN_VERSION_COUNT = 5

func main() {
	c := cron.New(cron.WithSeconds())                     //创建一个新的cron实例
	_, e := c.AddFunc("0 0 4 * * *", delObsoleteMetadata) //每天的凌晨4点执行定时任务
	if e != nil {
		log.Println(e)
		return
	}
	c.Start() //开始定时任务

	select {}
}

// delObsoleteMetadata 删除过期的元数据
func delObsoleteMetadata() {
	//搜索版本数量>=6的对象
	buckets, e := es.SearchVersionStatus(MIN_VERSION_COUNT + 1)
	if e != nil {
		log.Println(e)
		return
	}
	for i := range buckets {
		bucket := buckets[i]
		for v := 0; v < bucket.Doc_count-MIN_VERSION_COUNT; v++ {
			es.DelMetadata(bucket.Key, v+int(bucket.Min_version.Value))
		}
	}
}
