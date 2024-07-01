# OrionStore
基于Go语言的分布式对象存储系统  「Aliyun CentOS 7.9 64位」
   
## 项目背景
随着互联网、大数据和物联网技术的快速发展，非结构化数据（如图片、视频、日志文件等）占据了数据世界的大部分，对象存储为非结构化数据提供了有效的存储方案。分布式对象存储系统是一种专为处理大规模、非结构化数据而设计的存储解决方案。它是分布式系统技术与云计算发展到一定阶段的产物，旨在提供高可靠性、可扩展性、可用性和安全性的数据存储与管理服务。分布式对象存储系统通过将数据存储为独立的单元或“对象”，每个对象包括数据本身、可变数量的元数据以及全局唯一的标识符，从而实现对数据的有效组织和管理。

本项目主要实现的功能如下所示：
* 将RabbitMQ作为接口服务层和数据服务层之间的消息中间件，利于系统模块的解耦及扩展。
* 使用ElasticSearch对存储对象的元数据版本信息进行有效管理，并提升元数据的搜索效率。
* 用OpenSSL和Base64对数据进行校验和去重，确保数据准确性，有效节省存储空间。
* 运用Reed-Solomon Code(RS纠删码)对存储对象出现的错误进行即时修复。
* 对上传的对象数据进行Gzip压缩，且下载时自动解压缩，在用户无感的情况下，进一步节约存储空间。用户亦可选择以Gzip的方式进行下载，减少网络带宽及流量的消耗。
* 支持存储对象的分段上传及下载时的断点续传。
* 制定对象的版本留存策略，并且定期检查和修复损坏的对象数据。


## OrionStore分布式对象存储系统框架
<div align="middle">
<img alt="OrionStore" src="https://github.com/ZealACMer/OrionStore/assets/16794553/8ab7823d-5acf-4b71-a5e0-9ab83f18e9f0">
</div>
                        
## 依赖包
#### reedsolomon v1.12.1
```bash
go get github.com/klauspost/reedsolomon v1.12.1
```

#### amqp091-go v1.10.0
```bash
go get github.com/rabbitmq/amqp091-go v1.10.0
```

#### cron/v3 v3.0.1
```bash
go get github.com/robfig/cron/v3 v3.0.1
```
## 环境配置
#### rabbitmq-server的安装 (本项目使用v3.10.2)
请参照[https://github.com/rabbitmq/rabbitmq-server](https://github.com/rabbitmq/rabbitmq-server)
```bash
# 启动服务
$ sudo systemctl start rabbitmq-server

# 下载rabbitmqadmin管理工具
$ sudo rabbitmq-plugins enable rabbitmq_management
$ wget localhost:15672/cli/rabbitmqadmin

# 创建heartBeat和location两个exchange
$ python3 rabbitmqadmin declare exchange name=heartBeat type=fanout
$ python3 rabbitmqadmin declare exchange name=location type=fanout

# 添加用户test，密码设置为test
$ sudo rabbitmqctl add_user test test

# 为test用户添加访问exchange的权限
$ sudo rabbitmqctl set_permissions -p / test ".*" ".*" ".*"
```

#### ElasticSearch的安装(本项目使用v8.7.1)
请参照[https://github.com/elastic/elasticsearch](https://github.com/elastic/elasticsearch)
```bash
# 启动服务
$ sudo systemctl start elasticsearch

# 创建metadata索引及objects类型的映射
$ curl elasticsearch主机ip地址:9200/metadata -XPUT -d' {"mappings":{"objects":{"properties":{"name":{"type":"string","index":"not analyzed"},"version":{"type":"integer"},"size":{"type":"integer"},"hash":{"type":"string"}}}}}'
```

## 使用示例
- 本项目使用了4+2RS纠删码的模式，所以一共需要六台主机存储数据(存储4个数据分片+2个校验分片)，还需要两台主机承担接口服务器的角色，此外还需要主机分别用来部署RabbitMQ和ElasticSearch。在硬件资源受限的情况下，可以使用本机的网卡绑定8个虚拟网络地址(具体的设置，可以根据自己的操作系统，搜索相应的文档)，如此一来，就可以用一台机器同时模拟8台主机(6台存储数据的主机及2台用于接口服务的主机)的操作，RabbitMQ和ElasticSearch也可部署在本机上。
- 首先，在本机创建相应的文件夹，用于存储相关数据，之所以在tmp目录下创建1～6文件夹，是用来区分6个数据节点的存储目录，如果每个数据节点单独部署，可以省略此数字。
  ```bash
  #!/bin/bash
   for i in `seq 1 6`
   do
       mkdir -p /tmp/$i/objects
       mkdir -p /tmp/$i/temp
       mkdir -p /tmp/$i/garbage
   done
  ```
- 启动RabbitMQ和ElasticSearch之后，执行如下命令，用来模拟启动6个数据节点和2个接口服务节点。
   ```bash
   #!/bin/bash
   export RABBITMQ_SERVER=amqp://test:test@(主机ip地址):5672
   export ES_SERVER=(主机ip地址):9200
   
   LISTEN_ADDRESS=(数据节点1绑定的虚拟网络地址:端口号) STORAGE_ROOT=/tmp/1 go run dataServer/dataServer.go &
   LISTEN_ADDRESS=(数据节点2绑定的虚拟网络地址:端口号) STORAGE_ROOT=/tmp/2 go run dataServer/dataServer.go &
   LISTEN_ADDRESS=(数据节点3绑定的虚拟网络地址:端口号) STORAGE_ROOT=/tmp/3 go run dataServer/dataServer.go &
   LISTEN_ADDRESS=(数据节点4绑定的虚拟网络地址:端口号) STORAGE_ROOT=/tmp/4 go run dataServer/dataServer.go &
   LISTEN_ADDRESS=(数据节点5绑定的虚拟网络地址:端口号) STORAGE_ROOT=/tmp/5 go run dataServer/dataServer.go &
   LISTEN_ADDRESS=(数据节点6绑定的虚拟网络地址:端口号) STORAGE_ROOT=/tmp/6 go run dataServer/dataServer.go &
   
   LISTEN_ADDRESS=(接口服务节点1绑定的虚拟网络地址:端口号) go run apiServer/apiServer.go &
   LISTEN_ADDRESS=(接口服务节点2绑定的虚拟网络地址:端口号) go run apiServer/apiServer.go &
   ```
- 上传数据
  ```bash
  #获取上传对象的散列值
  openssl dgst -sha256 -binary /upload/1.mp3 | base64
  #进行对象的上传操作
  curl -v --data-binary @/upload/1.mp3 -X PUT (接口服务节点1或2绑定的虚拟网络地址:端口号)/objects/1.mp3 -H "digest: SHA-256=上传对象的散列值"
  ```
- 下载数据
  ```bash
  #非压缩格式下载
  curl -v (接口服务节点1或2绑定的虚拟网络地址:端口号)/objects/1.mp3 -o 1.mp3
  #以Gzip压缩格式下载
  curl -v (接口服务节点1或2绑定的虚拟网络地址:端口号)/objects/1.mp3 -o 1_mp3.gz -H "accept-encoding: gzip"
  #进行解压
  gunzip 1_mp3.gz
  ```
- 分段上传
  ```bash
  #获取文件的字节大小
  du -b /upload/2.MP4
  #获取文件的散列值
  openssl dgst -sha256 -binary /upload/2.MP4 | base64
  #获取上传文件的token
  curl -v (接口服务节点1或2绑定的虚拟网络地址:端口号)/objects/2.MP4 -X POST -H "digest: SHA-256=文件的散列值" -H "size:文件大小的字节数"
  #假设2.MP4的大小为1GB，上传第一段(设置为约640M)
  dd if=/upload/2.MP4 of=- bs=1000000 count=640 | 
  curl -v --data-binary @- -X PUT  (接口服务节点1或2绑定的虚拟网络地址:端口号)/(文件的token)
  #上传第二段
  dd if=/upload/2.MP4 of=- bs=1000000 skip=640 | 
  curl -v --data-binary @- -X PUT  (接口服务节点1或2绑定的虚拟网络地址:端口号)/(文件的token) -H "range: bytes=640000000-"
  ```
- 断点下载(从32000字节开始续传)
  ```bash
  curl -v -C -o 2.MP4 (接口服务节点1或2绑定的虚拟网络地址:端口号)/objects/2.MP4 -H "range: bytes=32000-"
  ```
- 操作对象的元数据
  ```bash
  #获取对象元数据的最新版本
  curl (接口服务节点1或2绑定的虚拟网络地址:端口号)/objects/1.jpg
  #获取对象元数据的指定版本
  curl (接口服务节点1或2绑定的虚拟网络地址:端口号)/objects/1.jpg?version=1 
  #获取对象元数据的所有版本
  curl (接口服务节点1或2绑定的虚拟网络地址:端口号)/versions/1.jpg
  #删除对象的元数据
  curl (接口服务节点1或2绑定的虚拟网络地址:端口号)/objects/1.jpg -XDELETE
  ```
- 根据对象版本的留存策略，删除陈旧的对象版本；删除数据节点的无效数据；以及数据节点中数据的自动检测与修复，系统后台会自动执行，无需手动操作
