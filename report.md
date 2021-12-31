## MongoDB log message 唯一消息统计
通过分析mongodb日志的唯一消息，可以直观的分析mongodb log message的信息
```bash
114514 WiredTiger message
 702 Slow query
 118 Connection ended
 111 client metadata
 111 Connection accepted
  29 Interrupted operation as its client disconnected
   5 Dropping all pooled connections
   5 Connecting
   2 Index build: done building
   2 Error sending response to client. Ending connection from remote

```
## 客户端类型分析
以下示例分析报告的远程[MongoDB 驱动程序](https://docs.mongodb.com/ecosystem/drivers/) 连接和客户端应用程序的[客户端数据](https://docs.mongodb.com/manual/reference/log-messages/#std-label-log-messages-client-data)，包括，并打印连接的每个唯一操作系统类型的总数，按频率排序
```bash
  64 linux
  21 Windows_NT
  14 Darwin
  12 Linux

```
## Application MongoDB驱动程序分析
以下为计算得到的所有远程[MongoDB 驱动程序](https://docs.mongodb.com/ecosystem/drivers/)连接数，并按出现次数降序显示每个驱动程序类型和版本
```bash
  64 {"name":"mongo-go-driver","version":"v1.7.1"}
  35 {"name":"nodejs","version":"4.1.4"}
   6 {"name":"NetworkInterfaceTL","version":"4.4.6"}
   6 {"name":"MongoDB Internal Client","version":"4.4.6"}

```
## Application Client分析
以下为计算得到的应用程序客户端的连接情况统计，并按出现次数降序显示每个驱动程序的类型和版本
```bash
 142 172.18.0.1
  84 10.10.30.17
  65 10.10.108.174
  42 10.10.108.40
   9 127.0.0.1

```
## 慢查询分析
慢查询记录了执行缓慢的语句，通过分析慢查询的规律来找到最需要优化的内容
```bash
{'time': '2021-11-29T11:57:56.852+08:00', 'insert': 'student', 'ns': 'mock.student', 'durationMillis': 114, 'storage': {}}
{'time': '2021-11-29T11:58:02.800+08:00', 'q': {}, 'u': {'$set': {'name': 'AAA'}}, 'ns': 'mock.student', 'durationMillis': 25791, 'planSummary': 'COLLSCAN', 'docsExamined': 151295, 'storage': {'data': {'bytesRead': 2, 'bytesWritten': 10648425, 'timeWritingMicros': 7995}}, 'appName': 'MongoDB Shell'}
{'time': '2021-11-29T11:58:02.801+08:00', 'update': 'student', 'ns': 'mock.$cmd', 'durationMillis': 25792, 'storage': {}, 'appName': 'MongoDB Shell'}
{'time': '2021-11-29T19:20:09.547+08:00', 'ns': 'admin.$cmd', 'durationMillis': 118, 'storage': {}, 'appName': 'MongoDB Shell'}
{'time': '2021-11-30T15:01:27.350+08:00', 'update': 'system.sessions', 'ns': 'config.$cmd', 'durationMillis': 145, 'storage': {}}

```
