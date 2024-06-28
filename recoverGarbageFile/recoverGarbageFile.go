package recoverGarbageFile

import (
	"github.com/robfig/cron/v3"
	"log"
	"orionStore/dataServer/locate"
	"orionStore/src/lib/es"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	c := cron.New(cron.WithSeconds())                          //创建一个新的cron实例
	_, e := c.AddFunc("0 20 4 * * *", startRecoverGarbageFile) //每天的凌晨4点20分执行定时任务
	if e != nil {
		log.Println(e)
		return
	}
	c.Start() //开始定时任务

	select {}
}

func startRecoverGarbageFile() {
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/garbage/*")
	for i := range files {
		file := strings.Split(filepath.Base(files[i]), ".")
		if len(file) != 3 {
			panic(files[i])
		}
		hash := file[0]
		id, e := strconv.Atoi(file[1])
		if e != nil {
			panic(e)
		}
		hashInMetadata, e := es.HasHash(hash)
		if e != nil {
			log.Println(e)
			return
		}
		if hashInMetadata {
			locate.Objects[hash] = id
			os.Rename(files[i], os.Getenv("STORAGE_ROOT")+"/objects/"+filepath.Base(files[0]))
		}
	}
}
