package heartbeat

import (
	"orionStore/src/lib/rabbitmq"
	"os"
	"time"
)

func StartHeartbeat() {
	//调用rabbitmq.New创建了一个rabbitmq.RabbitMQ结构体
	q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	defer q.Close()
	for {
		//在无限循环中调用rabbitmq.RabbitMQ结构体的Publish方法向heartBeat exchange发送本节点的监听地址
		q.Publish("heartBeat", os.Getenv("LISTEN_ADDRESS"))
		time.Sleep(5 * time.Second)
	}
}
