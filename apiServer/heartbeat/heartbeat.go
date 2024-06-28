package heartbeat

import (
	"orionStore/src/lib/rabbitmq"
	"os"
	"strconv"
	"sync"
	"time"
)

// key：数据服务节点的监听地址，value：收到心跳消息的时间
var dataServers = make(map[string]time.Time)

var mutex sync.RWMutex

// ListenHeartbeat 监听每个来自数据服务节点的心跳消息
func ListenHeartbeat() {
	q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	defer q.Close()
	q.Bind("heartBeat")
	c := q.Consume()
	//启动一个goroutine执行removeExpiredDataServer函数
	go removeExpiredDataServer()
	for msg := range c {
		dataServer, e := strconv.Unquote(string(msg.Body))
		if e != nil {
			panic(e)
		}
		mutex.Lock()
		dataServers[dataServer] = time.Now()
		mutex.Unlock()
	}
}

// removeExpiredDataServer 每隔5s扫描一遍dataServers，并清除其中超过10s没收到心跳消息的数据服务节点
func removeExpiredDataServer() {
	for {
		time.Sleep(5 * time.Second)
		mutex.Lock()
		for s, t := range dataServers {
			if t.Add(10 * time.Second).Before(time.Now()) {
				delete(dataServers, s)
			}
		}
		mutex.Unlock()
	}
}

// GetDataServers 遍历dataServers并返回当前所有的数据服务节点
func GetDataServers() []string {
	mutex.RLock()
	defer mutex.RUnlock()
	ds := make([]string, 0)
	for s, _ := range dataServers {
		ds = append(ds, s)
	}
	return ds
}
