package AccessMongoDB

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"mongostatus/Output"
)

/*Obtain various inspection indicators of the database*/

type dbstatsCollector struct {
	ctx    context.Context
	client *mongo.Client
}

func (d *dbstatsCollector) Collect() (db []string) {
	dbNames, err := d.client.ListDatabaseNames(d.ctx, bson.M{})
	if err != nil {
		log.Println("Failed to get database names: %s", err)
		return
	}
	var Db []string
	for _, db := range dbNames {

		//var dbStats bson.M
		//cmd := bson.D{{Key: "dbStats", Value: 1}, {Key: "scale", Value: 1}}
		//r := d.client.Database(db).RunCommand(d.ctx, cmd)
		//err := r.Decode(&dbStats)
		//if err != nil {
		//	log.Printf("Failed to get $dbstats for database %s: %s", db, err)
		//	continue
		//}
		//log.Printf("$dbStats metrics for %s", db)
		Db = append(Db, db)
	}
	return Db
}

func newdbstatsCollector(c context.Context, cl *mongo.Client) *dbstatsCollector {
	return &dbstatsCollector{
		c, cl,
	}
}

func GetDBSTATS(c context.Context, client *mongo.Client) (db []string) {
	dbstats := newdbstatsCollector(c, client)
	dbslice := dbstats.Collect()
	return dbslice
}

func GetDBfullinfo(c context.Context, client *mongo.Client, dbname []string, fl, fi string) {
	Output.Writeins(_dbstatInfo, fi)
	msg := fmt.Sprintf("[DBSTATS 数据库的统计信息]")
	Output.DoResult(msg, fl)
	Output.Writeins(msg, fi)
	for _, v := range dbname {
		Output.Writeins("```txt", fi)
		if err := RunCommandDBstats(c, client, v, fl, fi); err != nil {
			log.Println(err.Error())
		}
		Output.Writeins("```", fi)
	}
}

func RunCommandDBstats(ctx context.Context, cl *mongo.Client, dbname string, fl, fi string) error {
	//cmd := bson.D{{Key: "dbstats", Value: "1"}}
	//if err := cl.Database(dbname).RunCommand(ctx,cmd ).Decode(&md); err != nil {
	//	return  err
	//}
	var dbStats DBSTATS
	cmd := bson.D{{Key: "dbStats", Value: 1}, {Key: "scale", Value: 1024000000}} // 1024000000 以G为单位

	r := cl.Database(dbname).RunCommand(ctx, cmd)
	err := r.Decode(&dbStats)
	if err != nil {
		log.Printf("Failed to get $dbstats for database %s: %s", dbname, err)
	}
	if dbStats.OK {
		msg := fmt.Sprintf("[数据库 : %s] 集合数量: %v,视图数量: %v,索引总数: %v\n整个数据库文档数量: %v , 文档存储的空间总和: %vG\n索引的空间总和: %vG,文档和索引分配的空间总和: %vG\n存储数据的文件系统上已用的磁盘容量的总大小: %vG\n存储数据的文件系统上所有磁盘容量的总大小: %vG", dbStats.DB, dbStats.COLLECTIONS, dbStats.VIEWS, dbStats.INDEXES, dbStats.OBJECTS, dbStats.STORAGESIZE, dbStats.INDEXESIZE, dbStats.TOTALSIZE, dbStats.FSUSEDSIZE, dbStats.FSTOTALSIZE)
		Output.DoResult(msg, fl)
		Output.Writeins(msg, fi)
	}
	return nil
}

type DBSTATS struct {
	DB          string  `bson:"db"`
	COLLECTIONS int     `bson:"collections"` // 数据库中的集合数。
	VIEWS       int     `bson:"views"`       // 数据库中的视图数。
	OBJECTS     int     `bson:"objects"`     // 数据库中所有集合的对象（特别是文档）的数量。
	AVGOBJSIZE  float64 `bson:"avgObjSize"`  // 每个文档的平均大小（以字节为单位）。这是 dataSize除以文档数。
	DATASIZE    float64 `bson:"dataSize"`    // 数据库中保存的未压缩数据的总大小
	STORAGESIZE float64 `bson:"storageSize"` // 分配给数据库中所有集合用于文档存储的空间总和 ，包括可用空间。
	INDEXES     float64 `bson:"indexes"`     // 数据库中所有集合的索引总数。
	INDEXESIZE  float64 `bson"indexSize"`    // 分配给数据库中所有索引的空间总和，包括空闲索引空间。
	TOTALSIZE   float64 `bson:"totalSize"`   // 为数据库中所有集合中的文档和索引分配的空间总和。包括已用和可用的存储空间。这是的总和storageSize及 indexSize。
	//"scaleFactor": 1024000000, // 单位 G
	FSUSEDSIZE  float64 `bson:"fsUsedSize"`  // MongoDB 存储数据的文件系统上已用的磁盘容量的总大小。
	FSTOTALSIZE float64 `bson:"fsTotalSize"` // 单位 G  // MongoDB 存储数据的文件系统上所有磁盘容量的总大小。
	OK          bool    `bson:"ok"`
}
