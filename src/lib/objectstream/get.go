package objectstream

import (
	"fmt"
	"io"
	"net/http"
)

type GetStream struct {
	reader io.Reader
}

// newGetStream 输入url字符串，获取记录http响应正文的io.Reader
func newGetStream(url string) (*GetStream, error) {
	r, e := http.Get(url)
	if e != nil {
		return nil, e
	}
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dataServer return http code %d", r.StatusCode)
	}
	return &GetStream{r.Body}, nil
}

// NewGetStream 提供数据节点的地址和对象名用于读取对象。
func NewGetStream(server, object string) (*GetStream, error) {
	if server == "" || object == "" {
		return nil, fmt.Errorf("invalid server %s object %s", server, object)
	}
	return newGetStream("http://" + server + "/objects/" + object)
}

// GetStream.Read 只要实现了该方法，GetStream结构体就实现了io.Reader接口
func (r *GetStream) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}
