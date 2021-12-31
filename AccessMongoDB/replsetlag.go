package AccessMongoDB

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"mongostatus/dateformat"
	"mongostatus/mgodb"
	"time"
)

/* 统计RS的主要节点和次要节点的复制延迟 */
/* 实现不了rs.printReplicationInfo()命令 */

var oplog struct {
	OP   string      `bson:"op"`
	NS   string      `bson:"ns"`
	O    interface{} `bson:"o"`
	TS   interface{} `bson:"ts"`
	WALL interface{} `bson:"wall"`
	V    interface{} `bson:"v"`
}

func Getoplogwin(ctx context.Context, client *mongo.Client) {
	//msg := oplog
	//err := oplog.CollectionDocuments(ctx, client, 0, 1, 1).Decode(&msg)
	//if err != nil {
	//	log.Println(err)
	//
	//}
	var firstOplogDate, lastOplogDate string
	var firstOplogUnixTs ,lastOplogUnixTs int
	/* get first oplog Datetime */
	firstOplogDate,firstOplogUnixTs=getOplogDate(ctx,client,1,"$natural")   // 这里ts --> 改成$natural 更快 db.colleation.find().sort('$natural',1) - 根据自然集排序进行数据查询
	lastOplogDate,lastOplogUnixTs=getOplogDate(ctx,client,-1,"$natural")    // 这里ts --> 改成$natural 更快
	memUptimeFormatDay, memUptimeFormatHour, memUptimeFormatMin, memUptimeFormatSec :=dateformat.ResolveTime(lastOplogUnixTs-firstOplogUnixTs)
    log.Printf("** [opLog.rs] First OplogDate(Location: ASIA/Shanghai) : %s\n",firstOplogDate)
	log.Printf("** [opLog.rs] Last  OplogDate(Location: ASIA/Shanghai) : %s\n",lastOplogDate)
	log.Printf("** [opLog.rs] Oplog Window : %v Days %v Hours %v Mins %v Secs",memUptimeFormatDay,memUptimeFormatHour,memUptimeFormatMin,memUptimeFormatSec)

}


func getOplogDate(ctx context.Context,client *mongo.Client,sort int,sortkey string)(logDate string,unixTs int){
	var oplogDate string
	var ts2unix int64
	oplog := mgodb.NewMgo("local", "oplog.rs")
	cur := oplog.CollectionDocuments(ctx, client, 0, 1, sort, sortkey)

	for cur.Next(context.Background()) {
		// To decode into a struct, use cursor.Decode()
		// 上方limit1,只会返回一条数据
		result := struct {
			Foo string
			Bar int32
		}{}
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		// do something with result...
		// To get the raw bson bytes use cursor.Current
		raw := cur.Current
		// 获得Numberlong的时间戳 然后将Numberlong转Unix时间戳，最后转Date
		tsFirst := raw.Lookup("wall")
		ts2unix = tsFirst.DateTime() / 1000 // 这里直接取整即可 得到的就是unix时间戳
		tm := time.Unix(ts2unix, 0)
		oplogDate = tm.Format("2006-01-02 15:04:05") //得到的是ASIA/Shanghai的时间 与数据库中的比要多8小时
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	return oplogDate,int(ts2unix)

}