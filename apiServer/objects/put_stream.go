package objects

import (
	"fmt"
	"orionStore/apiServer/heartbeat"
	"orionStore/src/lib/rs"
)

func putStream(hash string, size int64) (*rs.RSPutStream, error) {
	//获取随机数据节点地址
	servers := heartbeat.ChooseRandomDataServers(rs.ALL_SHARDS, nil)
	if len(servers) != rs.ALL_SHARDS {
		return nil, fmt.Errorf("cannot find enough dataServer")
	}
	//调用rs.NewRSPutStream生成一个数据流
	return rs.NewRSPutStream(servers, hash, size)
}
