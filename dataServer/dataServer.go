package main

import (
	"log"
	"net/http"
	"orionStore/dataServer/heartbeat"
	"orionStore/dataServer/locate"
	"orionStore/dataServer/objects"
	"orionStore/dataServer/temp"
	"os"
)

func main() {
	locate.CollectObjects()
	//启动一个goroutine执行heartbeat.StartHeartbeat函数
	//由于该函数在一个goroutine中执行，所以就算不返回也不会影响其他功能
	go heartbeat.StartHeartbeat()
	//启动一个goroutine执行locate.StartLocate函数
	go locate.StartLocate()
	http.HandleFunc("/objects/", objects.Handler)
	http.HandleFunc("/temp/", temp.Handler)
	//处理函数注册成功后，调用http.ListenAndServe开始监听端口，该端口由系统环境变量LISTEN_ADDRESS定义
	//正常情况下，该函数始终监听端口上的请求，除非进程被信号中断；异常情况下，函数返回错误，log.Fatal打印错误信息并退出程序
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))
}
