package objects

import (
	"fmt"
	"orionStore/apiServer/heartbeat"
	"orionStore/apiServer/locate"
	"orionStore/src/lib/rs"
)

// GetStream 根据对象的散列值hash定位对象
func GetStream(hash string, size int64) (*rs.RSGetStream, error) {
	//调用locate.Locate定位对象
	locateInfo := locate.Locate(hash)
	//locateInfo长度小于4，返回定位失败的错误
	if len(locateInfo) < rs.DATA_SHARDS {
		return nil, fmt.Errorf("object %s locate fail, result %v", hash, locateInfo)
	}
	dataServers := make([]string, 0)
	//如果locateInfo长度不为6，获取用于接收恢复分片的数据节点
	if len(locateInfo) != rs.ALL_SHARDS {
		dataServers = heartbeat.ChooseRandomDataServers(rs.ALL_SHARDS-len(locateInfo), locateInfo)
	}
	return rs.NewRSGetStream(locateInfo, dataServers, hash, size)
}
