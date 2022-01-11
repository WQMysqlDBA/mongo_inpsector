# 使用方法
> 如下是该巡检工具的使用方法

## 工具介绍 
### 该巡检工具有两部分
- 巡检MongoDB数据库的状态信息
- 分析MongoDB数据库的log message，统计log message记录的一些指标

在物理机或者虚拟机跟QFusion的使用场景稍有不同
```text
1、QFusion 部署的MongDB mongodb内部认证如果用的headless
2、Docker部署的MongDB 如果走的是docker的网络
3、物理机或者虚拟机部署的MongoDB 通过内网进行数据同步和通信
复制集在通信的时候依赖内部这个网络
本工具采用的go-mongo-driver，在访问数据库节点的时候，需要通过连接的节点进去之后，访问内部网络然后取得各个节点的信息
如果用主机名等就回导致你的巡检工具报错网络错误

为避免这个一问题，需要将副本集内部的配置信息，如主机名等跟你的巡检工具所在的机器保持网络通信
所以在使用过程中：
对于QFusion的巡检，可以随便新建一个POD，什么类型都好（MySQL MongoDB etc..）然后进行巡检
```
### 工具使用
./mongostatus有这两种方式
```text
Usage of ./mongostatus:
  -conn string
        host
  -p string
        password
  -rs string
        rs name
  -u string
        user
```
```text
执行./mongostatus 
-conn为mongodb地址，组成为ip:port
-p 为秘密 无密码可以不指定此参数
-u 为账号 无账号可以不指定此参数
-rs 为复制集名称，如果你想连接到复制集，则需要指定复制集名称 你可以将复制集的所有连接地址都写上
```
例子
```text
./mongostatus -conn 10.10.30.17:30001,10.10.30.17:30002,10.10.30.17:30003 -rs rstest
在程序的stdout或者日志中，会打印出mongodb的连接地址
[URI] mongodb uri: mongodb://10.10.30.17:30001,10.10.30.17:30002,10.10.30.17:30003/?compressors=disabled&gssapiServiceName=mongodb&replicaSet=rstest
```

## 生成报告说明:
该工具执行完成之后，会在当前目录下生成一个inspector的目录
一个实例会生成`实例名称+inspector.log `,`实例名称+.md`的文件
```text
log文件为巡检的日志，后续用此文件来做统一的分析
.md文件为该实例的单独的巡检报告
```
