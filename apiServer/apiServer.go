package main

import (
	"log"
	"net/http"
	"orionStore/apiServer/heartbeat"
	"orionStore/apiServer/locate"
	"orionStore/apiServer/objects"
	"orionStore/apiServer/temp"
	"orionStore/apiServer/versions"
	"os"
)

func main() {
	//启动一个goroutine执行heartbeat.ListenHeartbeat函数
	go heartbeat.ListenHeartbeat()
	//objects.Handler处理URL以/objects/开头的对象请求
	http.HandleFunc("/objects/", objects.Handler)
	http.HandleFunc("/temp/", temp.Handler)
	//locate.Handler函数处理URL以/locate/开头的定位请求
	http.HandleFunc("/locate/", locate.Handler)
	//versions包的Handler函数用于处理URL以/versions/开头的请求
	http.HandleFunc("/versions/", versions.Handler)
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))
}
