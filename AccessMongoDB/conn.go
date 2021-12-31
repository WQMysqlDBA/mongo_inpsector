package AccessMongoDB

import (
	"context"
	"flag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

var mongo_conn_uri string

func ConnMongo() {
	var mongohost, mongouser, mongopasswd, rsname string
	flag.StringVar(&mongohost, "conn", "", "host")
	flag.StringVar(&mongouser, "u", "", "user")
	flag.StringVar(&mongopasswd, "p", "", "password")
	flag.StringVar(&rsname, "rs", "", "rs name")
	flag.Parse()
	//var isRs bool
	if rsname == "" && mongouser == "" {
		mongo_conn_uri = "mongodb://" + mongohost + "/?compressors=disabled&gssapiServiceName=mongodb"
	} else if rsname != "" && mongouser == "" {
		mongo_conn_uri = "mongodb://" + mongohost + "/?compressors=disabled&gssapiServiceName=mongodb&replicaSet=" + rsname
		//isRs = true
	} else if rsname == "" && mongouser != "" {
		mongo_conn_uri = "mongodb://" + mongouser + ":" + mongopasswd + "@" + mongohost + "/?authSource=admin&compressors=disabled&gssapiServiceName=mongodb"
	} else {
		mongo_conn_uri = "mongodb://" + mongouser + ":" + mongopasswd + "@" + mongohost + "/?authSource=admin&compressors=disabled&gssapiServiceName=mongodb&replicaSet=" + rsname
	    //isRs =true
	}
	// Atlas的格式
	//mongo_conn_uri = "mongodb+srv://root:letsg0@atlas-ch.pmjvs.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"
	log.Printf("** [URI] mongodb uri: %v", mongo_conn_uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongo_conn_uri))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// 连接不上就直接defer一个panci
	// Ping the primary
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err)
	} else {
		log.Print("** [NetWork] Sucessfully ping Primary node\n")
	}
	if ok := getisOnlyonemongo(rsname); ok {
		if err := client.Ping(ctx, readpref.Secondary()); err != nil {
			log.Println(err)
		} else {
			log.Print("** [NetWork] Sucessfully ping secondary node\n")
		}
	}
	/* get db name,input to _dbname[]string */
	_dbname := GetDBSTATS(ctx, client)
	var sysdb, nonsysdb []string
	for _, v := range _dbname {
		if v == "admin" || v == "local" || v == "config" {
			sysdb = append(sysdb, v)
		} else {
			nonsysdb = append(nonsysdb, v)
		}
	}
	log.Printf("** [DBinfo] sys database: %s ", sysdb)
	log.Printf("** [DBinfo] nosys database: %s ", nonsysdb)
	GetDBfullinfo(ctx,client,_dbname)
	for _, v := range _dbname {
		DBIndex(ctx, client, v)
	}
	DiagnosticData(ctx, client)
	//GetAsserts(ctx,client)
	//if isRs{
	//	ReplSetStatus(ctx, client)
	//}
	ReplSetStatus(ctx, client)
	Getoplogwin(ctx, client)


}




func getisOnlyonemongo(f string) bool {
	// 单节点 or mongos 不支持Ping secondary，因此会导致ctx超时，后面的代码跑不动
	// 所以在执行这个之前，判断下 rs参数是否有填
	//    不填写表示为 单节点/mongos
	//    但是也有可能是副本集只暴露主节点 也是ok的
	if f == "" {
		// 为空不检查
		return false
	} else {
		return true
	}
}
