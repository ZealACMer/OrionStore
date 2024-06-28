package types

// LocateMessage 由于该结构体同时被接口服务和数据服务引用，所以放在types包里
type LocateMessage struct {
	Addr string //存放分片的数据节点的地址
	Id   int    //分片的id
}
