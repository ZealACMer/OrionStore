package locate

import (
	"encoding/json"
	"orionStore/src/lib/rabbitmq"
	"orionStore/src/lib/rs"
	"orionStore/src/lib/types"
	"os"
	"time"
)

// Locate 参数name：需要定位的对象的名字 返回值：map(key:分片id，value：存储分片的数据节点的地址)
func Locate(name string) (locateInfo map[int]string) {
	//创建一个新的消息队列
	q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	//向location exchange群发对象名字的定位信息
	q.Publish("location", name)
	c := q.Consume()
	//使用用goroutine启动一个匿名函数，设置超时机制，用于在1s后关闭这个消息队列
	go func() {
		time.Sleep(time.Second)
		q.Close()
	}()
	locateInfo = make(map[int]string)
	//rs.ALL_SHARDS为常数6，代表一共有4+2个分片
	for i := 0; i < rs.ALL_SHARDS; i++ {
		msg := <-c
		//如果收到一个长度为0的消息，则返回一个空的map
		if len(msg.Body) == 0 {
			return
		}
		var info types.LocateMessage
		json.Unmarshal(msg.Body, &info)
		//key:分片id，value：存储分片的数据服务节点地址
		locateInfo[info.Id] = info.Addr
	}
	return
}

// Exist 参数：需要定位的对象的名字，逻辑：判断收到的反馈信息数量是否>=4，为true则说明对象存在。
func Exist(name string) bool {
	return len(Locate(name)) >= rs.DATA_SHARDS
}
