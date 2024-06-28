package locate

import (
	"orionStore/src/lib/rabbitmq"
	"orionStore/src/lib/types"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Objects 用于缓存所有的对象信息
var Objects = make(map[string]int)

// mutex 读写锁用于保护对objects的读写操作
var mutex sync.RWMutex

// Locate 定位对象，并返回对象分片的id
func Locate(hash string) int {
	mutex.RLock()
	id, ok := Objects[hash]
	mutex.RUnlock()
	if !ok {
		return -1
	}
	return id
}

// Add 用于将对象及其分片的id加入缓存
func Add(hash string, id int) {
	mutex.Lock()
	Objects[hash] = id
	mutex.Unlock()
}

// Del 用于将一个散列值在缓存中移除
func Del(hash string) {
	mutex.Lock()
	delete(Objects, hash)
	mutex.Unlock()
}

// StartLocate 用于监听来自于接口服务的定位消息
func StartLocate() {
	// 创建一个rabbitmq.RabbitMQ的结构体
	q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	defer q.Close()
	// 调用Bind方法绑定location exchange
	q.Bind("location")
	//调用Consume方法返回一个channel
	c := q.Consume()
	for msg := range c {
		hash, e := strconv.Unquote(string(msg.Body))
		if e != nil {
			panic(e)
		}
		//调用Locate函数检查对象是否存在
		id := Locate(hash)
		if id != -1 {
			//如果存在，则调用Send方法向消息的发送方返回本服务节点的监听地址，表示该对象存在于本服务节点上
			q.Send(msg.ReplyTo, types.LocateMessage{Addr: os.Getenv("LISTEN_ADDRESS"), Id: id})
		}
	}
}

// CollectObjects 更新定位缓存信息
func CollectObjects() {
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/*")
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
		Objects[hash] = id
	}
}
