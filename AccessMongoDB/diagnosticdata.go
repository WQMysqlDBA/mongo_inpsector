package AccessMongoDB

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"mongostatus/Output"
	"mongostatus/dateformat"
	"mongostatus/util"
)

type diagnosticDataCollector struct {
	ctx            context.Context
	client         *mongo.Client
	compatibleMode bool
	topologyInfo   labelsGetter
}

func (d *diagnosticDataCollector) Collect(fl, fi string) (ch chan bool) {
	var m bson.M

	cmd := bson.D{{Key: "getDiagnosticData", Value: "1"}}
	res := d.client.Database("admin").RunCommand(d.ctx, cmd)
	if res.Err() != nil {
		if isArbiter, _ := isArbiter(d.ctx, d.client); isArbiter {
			return
		}
	}

	if err := res.Decode(&m); err != nil {
		msg := fmt.Sprintf("cannot run getDiagnosticData: %s", err)
		Output.DoResult(msg, fl)
	}

	m, ok := m["data"].(bson.M)
	if !ok {
		err := fmt.Sprintf("%T for data field", m["data"])
		msg := fmt.Sprintf("cannot decode getDiagnosticData: %s", err)
		Output.DoResult(msg, fl)
	}

	Output.DoResult(" ** [getDiagnosticData]", fl)

	service, err := util.GetServerStatus(d.ctx, d.client)
	if err != nil {
		log.Println(err.Error())
	}

	// baseinfo
	Output.Writeins(_baseInfo, fi)
	Output.Writeins("```txt", fi)
	// 主机名
	msg := fmt.Sprintf(" ** [HostName] : %s", service.Host)
	Output.DoResult(msg, fl)
	// 数据库版本
	msg = fmt.Sprintf(" ** [Mongod Version] : %s", service.Version)
	Output.DoResult(msg, fl)
	// 数据库Processname和PID
	msg = fmt.Sprintf(" ** [Mongod Process] : processname : %s, pid : %v", service.Process, service.Pid)
	Output.DoResult(msg, fl)
	// 数据库启动运行时间
	day, hour, min, sec := dateformat.ResolveTime(int(service.Uptime))
	msg = fmt.Sprintf(" ** [Mongod Uptime] : %v Days %v Hours %v Mins %v secs", day, hour, min, sec)
	Output.DoResult(msg, fl)
	msg = fmt.Sprintf("[Hostname] : %s\n[Mongod Version] : %s\n[Mongod Process name] : %s\n[Mongod Process pid] : %v", service.Host, service.Version, service.Process, service.Pid)
	Output.Writeins(msg, fi)
	// 判断节点类型  是单节点还是复制集节点还是mongos
	nodeType, err := util.GetNodeType(d.ctx, d.client)
	if err != nil {
		m := fmt.Sprintf("[Unknow] Cannot get node type to check if this is a mongos: %s", err)
		Output.DoResult(m, fl)
	} else {
		m := fmt.Sprintf("[architecture]: %s", nodeType)
		Output.DoResult(m, fl)
		Output.Writeins(m, fi)
	}
	Output.Writeins("```", fi)

	// 内存 当前内存使用情况的文档
	Output.Writeins(_memoryInfo, fi)
	Output.Writeins("```txt", fi)

	msg = fmt.Sprintf("[MongoDB MEMORY INFO]")
	Output.DoResult(msg, fl)
	msg = fmt.Sprintf("[MongoDB MEMORY] OS bits : %v", service.Mem.Bits)
	Output.DoResult(msg, fl)
	Output.Writeins(msg, fi)
	msg = fmt.Sprintf("[MongoDB MEMORY] Mem resident : %v (MiB) (该值大致相当于数据库进程当前使用的 RAM 量)", service.Mem.Resident)
	Output.DoResult(msg, fl)
	Output.Writeins(msg, fi)
	msg = fmt.Sprintf("[MongoDB MEMORY] Mem virtual : %v (MiB) (进程使用的虚拟内存的数量)", service.Mem.Virtual)
	Output.DoResult(msg, fl)
	Output.Writeins(msg, fi)

	if service.Mem.Supported.(bool) {
		msg = fmt.Sprintf("[MongoDB MEMORY] Mem supported : %v (底层系统是否支持扩展内存信息)", service.Mem.Supported)
		Output.DoResult(msg, fl)
		Output.Writeins(msg, fi)
	} else {
		msg = fmt.Sprintf("[MongoDB MEMORY] Mem supported : %v (底层系统是否支持扩展内存信息)", service.Mem.Supported)
		Output.DoResult(msg, fl)
		Output.Writeins(msg, fi)
		msg = fmt.Sprintf("[MongoDB MEMORY] Mem note : %v (mem.supported为假，则该字段出现)", service.Mem.Note)
		Output.DoResult(msg, fl)
		Output.Writeins(msg, fi)
	}

	Output.Writeins("```", fi)

	// 连接数情况
	Output.Writeins(_connectionInfo, fi)
	Output.Writeins("```txt", fi)
	msg = fmt.Sprintf("[MongoDB CONNECTIONS INFO]")
	Output.DoResult(msg, fl)
	Output.Writeins(msg, fi)
	msg = fmt.Sprintf("[MongoDB CONNECTIONS] TOTAL CREATED %v ", service.Connections.TotalCreated)
	Output.DoResult(msg, fl)
	Output.Writeins(msg, fi)
	msg = fmt.Sprintf("[MongoDB CONNECTIONS] Current %v ", service.Connections.Current)
	Output.DoResult(msg, fl)
	Output.Writeins(msg, fi)
	msg = fmt.Sprintf("[MongoDB CONNECTIONS] Available %v ", service.Connections.Available)
	Output.DoResult(msg, fl)
	Output.Writeins(msg, fi)
	Output.Writeins("```", fi)

	// 数据库存储引擎
	Output.Writeins(_engineInfo, fi)
	Output.Writeins("```txt", fi)
	msg = fmt.Sprintf("[MongoDB STORAGE Engine INFO]")
	Output.DoResult(msg, fl)
	Output.Writeins(msg, fi)
	msg = fmt.Sprintf("[MongoDB STORAGE] Engine Name : %s", service.StorageEngine.Name)
	Output.DoResult(msg, fl)
	Output.Writeins(msg, fi)
	msg = fmt.Sprintf("[MongoDB STORAGE] Engine Persistent : %v (该参数表示存储引擎是否支持持久化数据到硬盘)", service.StorageEngine.Persistent)
	Output.DoResult(msg, fl)
	Output.Writeins(msg, fi)
	msg = fmt.Sprintf("[MongoDB STORAGE] Engine SupportCommittedREads : %v (该参数表示存储引擎是否支持read concern)", service.StorageEngine.SupportsCommittedREads)
	Output.DoResult(msg, fl)
	Output.Writeins(msg, fi)
	Output.Writeins("```", fi)

	// 数据库断言 service.Asserts是一个map类型

	//  "regular" : 0,   //服务启动后正常的asserts错误个数,可通过log查看更多该信息
	//	"warning" : 0,   //服务启动后的warning个数
	//	"msg" : 0,     //服务启动后的message assert个数
	//	"user" : 0,    //服务启动后的user asserts个数
	//	"rollovers" : 0  //服务启动后的重置次数

	const (
		//_space       = "                       "
		_space       = ""
		_asserts_msg = _space + "\"regular\" : 服务启动后正常的asserts错误个数,可通过log查看更多该信息\n" +
			_space + "\"warning\" : 服务启动后的warning个数\n" +
			_space + "\"msg\" : 服务启动后的message assert个数\n" +
			_space + "\"user\" : 服务启动后的user asserts个数\n" +
			_space + "\"rollovers\" : 服务启动后的重置次数"
	)
	Output.Writeins(_assertsOnfo, fi)
	Output.Writeins("```txt", fi)
	msg = fmt.Sprintf("[Asserts INFO]")
	Output.Writeins(msg, fi)

	if len(service.Asserts) == 0 {
		Output.DoResult("[Warn] Asserts is nil", fl)
		Output.Writeins("[Warn] Asserts is nil", fi)
	} else {
		for k, v := range service.Asserts {
			m := fmt.Sprintf("[Asserts] %s : %v", k, v)
			Output.DoResult(m, fl)
			Output.Writeins(m, fi)
		}
	}
	Output.Writeins("```", fi)
	Output.Writeins("**[对于assert各项的说明]**\n", fi)
	Output.Writeins("```txt", fi)
	msg = fmt.Sprintf("%s", _asserts_msg)
	Output.DoResult(msg, fl)
	Output.Writeins(msg, fi)
	Output.Writeins("```", fi)

	// 事务计算TPS 可以做成监控，比如计算1s做了多少个trx 初始化一个容量为2的数组，index[0]为上一秒的值，index[1]为当前值
	// 启动这个模块的G1 ，（当前的totalCommitted - 上一个totalCommitted ）/ 时间间隔 --> tps : n trx/s ，之后将当前的totalCommitted传递给下一个gotine
	// 或者用channel也可以 channel只存上一次的totalCommitted值
	// wiredTiger.transaction

	// repl
	c := make(chan bool, 1)
	if service.Repl != nil {
		// 如果是单节点，db.RunCommandDBstats({serverStatus:1}).repl 是一个nil pointer，会导致panic
		// 加一层判断，防止panic
		//走到这一层，就已经是rs的配置了
		Output.Writeins(_replInfo, fi)
		m := fmt.Sprintf("** [Replset INFO]:")
		Output.DoResult(m, fl)
		Output.Writeins("```txt", fi)
		if service.Repl.IsMaster.(bool) {
			m := fmt.Sprintf("[Replset INFO] node role : %s", "PRIMARY")
			Output.DoResult(m, fl)
			Output.Writeins(m, fi)
		} else if service.Repl.Secondary.(bool) {
			m := fmt.Sprintf("[Replset INFO] node role : %s", "SECONDARY")
			Output.DoResult(m, fl)
			Output.Writeins(m, fi)
		}

		c <- true
	} else {
		c <- false
	}
	return c
}

func huanhang() {
	fmt.Println()
}

func newdiagnosticDataCollector(c context.Context, cl *mongo.Client) *diagnosticDataCollector {
	var a labelsGetter
	return &diagnosticDataCollector{
		c, cl, true, a,
	}
}

func DiagnosticData(c context.Context, client *mongo.Client, f, f1 string) (ch chan bool) {
	diadata := newdiagnosticDataCollector(c, client)
	return diadata.Collect(f, f1)
}
