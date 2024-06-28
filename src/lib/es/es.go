package es

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Metadata struct {
	Name    string
	Version int
	Size    int64
	Hash    string
}

type hit struct {
	Source Metadata `json:"_source"`
}

//type searchResult struct {
//	Hits struct {
//		Total int
//		Hits  []hit
//	}
//}

type totalHits struct {
	Value    int
	Relation string
}

type searchResult struct {
	Hits struct {
		Total totalHits
		Hits  []hit
	}
}

// getMetadata 根据对象的名字和版本号来获取对象的元数据
func getMetadata(name string, versionId int) (meta Metadata, e error) {
	url_ := fmt.Sprintf("http://%s/metadata/objects/%s_%d/_source",
		os.Getenv("ES_SERVER"), name, versionId)
	r, e := http.Get(url_)
	if e != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf("fail to get %s_%d: %d", name, versionId, r.StatusCode)
		return
	}
	result, _ := io.ReadAll(r.Body)
	json.Unmarshal(result, &meta)
	return
}

// SearchLatestVersion 以对象的名字为参数，调用ES搜索API，获得版本号以降序排列的第一个结果
func SearchLatestVersion(name string) (meta Metadata, e error) {
	url_ := fmt.Sprintf("http://%s/metadata/_search?q=name:%s&size=1&sort=version:desc",
		os.Getenv("ES_SERVER"), url.PathEscape(name))
	r, e := http.Get(url_)
	if e != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf("fail to search latest metadata: %d", r.StatusCode)
		return
	}
	result, _ := io.ReadAll(r.Body)
	var sr searchResult
	err := json.Unmarshal(result, &sr)
	if err != nil {
		log.Printf("JSON unmarshal error: %v", err)
		return
	}
	//如果ES返回的结果长度为O，说明没有搜到相对应的元数据，直接返回，此时meta中各属性都为初始值:字符串为""，整型为0
	if len(sr.Hits.Hits) != 0 {
		meta = sr.Hits.Hits[0].Source
	}
	return
}

// GetMetadata 输入对象的名字和版本号返回对象, 当version为0时，获取当前最新的版本
func GetMetadata(name string, version int) (Metadata, error) {
	if version == 0 {
		return SearchLatestVersion(name)
	}
	return getMetadata(name, version)
}

// PutMetadata 用于向ES服务上传一个新的元数据
func PutMetadata(name string, version int, size int64, hash string) error {
	doc := fmt.Sprintf(`{"name":"%s","version":%d,"size":%d,"hash":"%s"}`,
		name, version, size, hash)
	client := http.Client{}
	//op_type=create，如果同时有多个客户端上传同一个元数据，结果会发生冲突，只有第一个文档被成功创建
	url := fmt.Sprintf("http://%s/metadata/objects/%s_%d?op_type=create",
		os.Getenv("ES_SERVER"), name, version)
	request, _ := http.NewRequest("PUT", url, strings.NewReader(doc))
	r, e := client.Do(request)
	if e != nil {
		return e
	}
	//处理版本冲突
	if r.StatusCode == http.StatusConflict {
		return PutMetadata(name, version+1, size, hash)
	}
	if r.StatusCode != http.StatusCreated {
		result, _ := io.ReadAll(r.Body)
		return fmt.Errorf("fail to put metadata: %d %s", r.StatusCode, string(result))
	}
	return nil
}

// AddVersion 首先调用SearchLatestVersion获取对象最新的版本，然后在版本号上加1调用PutMetadata
func AddVersion(name, hash string, size int64) error {
	version, e := SearchLatestVersion(name)
	if e != nil {
		return e
	}
	return PutMetadata(name, version.Version+1, size, hash)
}

// SearchAllVersions 函数用于搜索某个对象或所有对象的全部版本
func SearchAllVersions(name string, from, size int) ([]Metadata, error) {
	url := fmt.Sprintf("http://%s/metadata/_search?sort=name,version&from=%d&size=%d",
		os.Getenv("ES_SERVER"), from, size)
	if name != "" {
		url += "&q=name:" + name
	}
	r, e := http.Get(url)
	if e != nil {
		return nil, e
	}
	metas := make([]Metadata, 0)
	result, _ := io.ReadAll(r.Body)
	var sr searchResult
	err := json.Unmarshal(result, &sr)
	if err != nil {
		log.Printf("JSON unmarshal error: %v", err)
		return metas, err
	}
	for i := range sr.Hits.Hits {
		metas = append(metas, sr.Hits.Hits[i].Source)
	}
	return metas, nil
}

// DelMetadata 根据对象的名字name和版本号version删除相应的对象元数据
func DelMetadata(name string, version int) {
	client := http.Client{}
	url := fmt.Sprintf("http://%s/metadata/objects/%s_%d",
		os.Getenv("ES_SERVER"), name, version)
	request, _ := http.NewRequest("DELETE", url, nil)
	client.Do(request)
}

type Bucket struct {
	Key         string //对象的名字
	Doc_count   int    //对象目前的版本个数
	Min_version struct {
		Value float32 //对象当前最小的版本号
	}
}

type aggregateResult struct {
	Aggregations struct {
		Group_by_name struct {
			Buckets []Bucket
		}
	}
}

// SearchVersionStatus 搜索版本数量>=min_doc_count的对象
func SearchVersionStatus(min_doc_count int) ([]Bucket, error) {
	client := http.Client{}
	url := fmt.Sprintf("http://%s/metadata/_search", os.Getenv("ES_SERVER"))
	body := fmt.Sprintf(`
        {
          "size": 0,
          "aggs": {
            "group_by_name": {
              "terms": {
                "field": "name",
                "min_doc_count": %d
              },
              "aggs": {
                "min_version": {
                  "min": {
                    "field": "version"
                  }
                }
              }
            }
          }
        }`, min_doc_count)
	request, _ := http.NewRequest("GET", url, strings.NewReader(body))
	r, e := client.Do(request)
	if e != nil {
		return nil, e
	}
	b, _ := io.ReadAll(r.Body)
	var ar aggregateResult
	json.Unmarshal(b, &ar)
	return ar.Aggregations.Group_by_name.Buckets, nil
}

// HasHash 搜索散列值是否在元数据服务中存在
func HasHash(hash string) (bool, error) {
	url := fmt.Sprintf("http://%s/metadata/_search?q=hash:%s&size=0", os.Getenv("ES_SERVER"), hash)
	r, e := http.Get(url)
	if e != nil {
		return false, e
	}
	b, _ := io.ReadAll(r.Body)
	var sr searchResult
	json.Unmarshal(b, &sr)
	return sr.Hits.Total.Value != 0, nil
}

// SearchHashSize 返回散列值对应的元数据属性中的size值
func SearchHashSize(hash string) (size int64, e error) {
	url := fmt.Sprintf("http://%s/metadata/_search?q=hash:%s&size=1",
		os.Getenv("ES_SERVER"), hash)
	r, e := http.Get(url)
	if e != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf("fail to search hash size: %d", r.StatusCode)
		return
	}
	result, _ := ioutil.ReadAll(r.Body)
	var sr searchResult
	json.Unmarshal(result, &sr)
	if len(sr.Hits.Hits) != 0 {
		size = sr.Hits.Hits[0].Source.Size
	}
	return
}
