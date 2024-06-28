package objectstream

import (
	"fmt"
	"io"
	"net/http"
)

type PutStream struct {
	writer *io.PipeWriter //用于实现Write方法
	c      chan error     //将goroutine传输数据过程中发生的错误传回主线程
}

// NewPutStream 用于生成一个PutStream结构体
func NewPutStream(server, object string) *PutStream {
	//用io.Pipe()创建了一对reader和writer，类型分别是*io.PipeReader和*io.PipeWriter，两者管道互联，
	//写入writer的内容可以从reader中读出来
	reader, writer := io.Pipe()
	c := make(chan error)
	go func() {
		request, _ := http.NewRequest("PUT", "http://"+server+"/objects/"+object, reader)
		client := http.Client{}
		r, e := client.Do(request)
		if e == nil && r.StatusCode != http.StatusOK {
			e = fmt.Errorf("dataServer return http code %d", r.StatusCode)
		}
		//将错误发送进channel c
		c <- e
	}()
	return &PutStream{writer, c}
}

// Write PutStream.Write方法用于写入writer，只有实现了这个方法，PutStream才会被认为是实现了io.Write接口
func (w *PutStream) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

// Close PutStream.Close方法用于关闭writer
func (w *PutStream) Close() error {
	w.writer.Close()
	return <-w.c
}
