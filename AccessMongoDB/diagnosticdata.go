package AccessMongoDB

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"mongostatus/dateformat"
	"mongostatus/util"
)

type diagnosticDataCollector struct {
	ctx            context.Context
	client         *mongo.Client
	compatibleMode bool
	topologyInfo   labelsGetter
}

func (d *diagnosticDataCollector) Collect() {
	var m bson.M

	cmd := bson.D{{Key: "getDiagnosticData", Value: "1"}}
	res := d.client.Database("admin").RunCommand(d.ctx, cmd)
	if res.Err() != nil {
		if isArbiter, _ := isArbiter(d.ctx, d.client); isArbiter {
			return
		}
	}

	if err := res.Decode(&m); err != nil {
		log.Printf("cannot run getDiagnosticData: %s", err)
	}

	m, ok := m["data"].(bson.M)
	if !ok {
		err := fmt.Sprintf("%T for data field", m["data"])
		log.Printf("cannot decode getDiagnosticData: %s", err)
	}

	log.Println("** [getDiagnosticData]")
	huanhang()

	service, err := util.GetServerStatus(d.ctx, d.client)
	if err != nil {
		log.Println(err.Error())
	}

	// 主机名
	log.Printf("** [HostName] : %s", service.Host)
	// 数据库版本
	log.Printf("** [Mongod Version] : %s", service.Version)
	// 数据库PID
	log.Printf("** [Mongod Process] : processname : %s, pid : %v", service.Process, service.Pid)

	// 数据库启动运行时间
	day, hour, min, sec := dateformat.ResolveTime(int(service.Uptime))
	log.Printf("** [Mongod Uptime] : %v Days %v Hours %v Mins %v secs", day, hour, min, sec)
	huanhang()

	// 内存 当前内存使用情况的文档
	log.Printf("** [MongoDB MEMORY INFO]")
	log.Printf("** [MongoDB MEMORY] OS bits : %v", service.Mem.Bits)
	log.Printf("** [MongoDB MEMORY] Mem resident : %v (MiB) (该值大致相当于数据库进程当前使用的 RAM 量)", service.Mem.Resident)
	log.Printf("** [MongoDB MEMORY] Mem virtual : %v (MiB) (进程使用的虚拟内存的数量)", service.Mem.Virtual)

	if service.Mem.Supported.(bool) {
		log.Printf("** [MongoDB MEMORY] Mem supported : %v (底层系统是否支持扩展内存信息)", service.Mem.Supported)
	} else {
		log.Printf("** [MongoDB MEMORY] Mem supported : %v (底层系统是否支持扩展内存信息)", service.Mem.Supported)
		log.Printf("** [MongoDB MEMORY] Mem note : %v (mem.supported为假，则该字段出现)", service.Mem.Note)
	}
	huanhang()

	// 连接数情况
	log.Printf("** [MongoDB CONNECTIONS INFO]")
	log.Printf("** [MongoDB CONNECTIONS] TOTAL CREATED %v ", service.Connections.TotalCreated)
	log.Printf("** [MongoDB CONNECTIONS] Current %v ", service.Connections.Current)
	log.Printf("** [MongoDB CONNECTIONS] Available %v ", service.Connections.Available)
	huanhang()

	// 数据库存储引擎
	log.Printf("** [MongoDB STORAGE Engine INFO]")
	log.Printf("** [MongoDB STORAGE] Engine Name : %s", service.StorageEngine.Name)
	log.Printf("** [MongoDB STORAGE] Engine Persistent : %v (该参数表示存储引擎是否支持持久化数据到硬盘)", service.StorageEngine.Persistent)
	log.Printf("** [MongoDB STORAGE] Engine SupportCommittedREads : %v (该参数表示存储引擎是否支持read concern)", service.StorageEngine.SupportsCommittedREads)
	huanhang()

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

	log.Printf("** [Asserts INFO]")
	if len(service.Asserts) == 0 {
		log.Printf("** [Warn] Asserts is nil")
	} else {
		for k, v := range service.Asserts {
			log.Printf("** [Asserts] %s : %v", k, v)
		}
	}
	log.Printf("** [对于assert各项的说明] : \n%s", _asserts_msg)
	huanhang()

	// 判断节点类型  是单节点还是复制集节点还是mongos
	nodeType, err := util.GetNodeType(d.ctx, d.client)
	if err != nil {
		log.Printf("** [Unknow] Cannot get node type to check if this is a mongos: %s", err)
	} else {
		log.Printf("** [architecture]: %s\n", nodeType)
	}
	huanhang()

	// 事务计算TPS 可以做成监控，比如计算1s做了多少个trx 初始化一个容量为2的数组，index[0]为上一秒的值，index[1]为当前值
	// 启动这个模块的G1 ，（当前的totalCommitted - 上一个totalCommitted ）/ 时间间隔 --> tps : n trx/s ，之后将当前的totalCommitted传递给下一个gotine
	// 或者用channel也可以 channel只存上一次的totalCommitted值
	// wiredTiger.transaction

	// repl
	if service.Repl != nil {
		// 如果是单节点，db.RunCommandDBstats({serverStatus:1}).repl 是一个nil pointer，会导致panic
		// 加一层判断，防止panic
		//走到这一层，就已经是rs的配置了
		log.Printf("** [Replset INFO]:")
		if service.Repl.IsMaster.(bool) {
			log.Printf("** [Replset INFO] node role : %s\n", "PRIMARY")
		} else if service.Repl.Secondary.(bool) {
			log.Printf("** [Replset INFO] node role : %s\n", "SECONDARY")
		}
	}
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

func DiagnosticData(c context.Context, client *mongo.Client) {
	diadata := newdiagnosticDataCollector(c, client)
	diadata.Collect()
}
