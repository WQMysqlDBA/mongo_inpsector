package AccessMongoDB

import (
	"context"
	"flag"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"mongostatus/Output"
	"strconv"
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
	log.Printf(" ** [URI] mongodb uri: %v", mongo_conn_uri)

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
		log.Print(" ** [NetWork] Sucessfully ping Primary node\n")
	}
	if ok := getisOnlyonemongo(rsname); ok {
		if err := client.Ping(ctx, readpref.Secondary()); err != nil {
			log.Println(err)
		} else {
			log.Print(" ** [NetWork] Sucessfully ping secondary node\n")
		}
	}

	// ============================================ Initlicate Insepctor File ============================================
	var msg string = ""
	f := mongohost + ".md"
	f1 := mongohost + "inspector.log"
	InspectorFileName, initerr := Output.InitInsepectorFile(f)
	if initerr != nil {
		panic(initerr.Error())
	}
	InsLog, initlogerr := Output.InitInsepectorFile(f1)
	if initlogerr != nil {
		panic(initlogerr.Error())
	}

	Output.Initresultfile(InspectorFileName)
	t := time.Now()
	year := t.Year()
	month := t.Month()
	day := t.Day()
	inspectorDate := strconv.Itoa(year) + "-" + strconv.Itoa(int(month)) + "-" + strconv.Itoa(day)
	headers := "# " + mongohost + _welComeInfo + "* 巡检日期\n" + inspectorDate
	Output.Writeins(headers, InspectorFileName)

	Output.Initresultfile(InsLog)

	msg = fmt.Sprintf(" ** [Info]: Write inspector result in file %s", InspectorFileName)
	Output.DoResult(msg, InsLog)

	msg = fmt.Sprintf(" ** [Info]: Write inspector log in file %s", InsLog)
	Output.DoResult(msg, InsLog)

	/* DBinfo */
	_dbname := GetDBSTATS(ctx, client)
	var sysdb, nonsysdb []string
	for _, v := range _dbname {
		if v == "admin" || v == "local" || v == "config" {
			sysdb = append(sysdb, v)
		} else {
			nonsysdb = append(nonsysdb, v)
		}
	}
	msg = fmt.Sprintf(" ** [DBinfo] sys database: %s", sysdb)
	Output.DoResult(msg, InsLog)
	msg = fmt.Sprintf(" ** [DBinfo] nosys database: %s", nonsysdb)
	Output.DoResult(msg, InsLog)

	// DiagnosticData
	isRepl := <-DiagnosticData(ctx, client, InsLog, InspectorFileName)

	if isRepl {
		ReplSetStatus(ctx, client, InsLog, InspectorFileName)
		Getoplogwin(ctx, client, InsLog, InspectorFileName)
	}

	GetDBfullinfo(ctx, client, _dbname, InsLog, InspectorFileName)

	// Indexinfo
	Output.Writeins(_indexCollect, InspectorFileName)
	for _, v := range _dbname {
		DBIndex(ctx, client, v, InsLog, InspectorFileName)
	}
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
