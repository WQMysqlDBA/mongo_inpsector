package AccessMongoDB

/* 统计复制集的状态和运行情况 */
import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"mongostatus/dateformat"
	"mongostatus/proto"
	"strconv"
	"time"
)

const (
	replicationNotEnabled        = 76
	replicationNotYetInitialized = 94
)

type replSetGetStatusCollector struct {
	ctx            context.Context
	client         *mongo.Client
	compatibleMode bool
	topologyInfo   labelsGetter
}

func (d *replSetGetStatusCollector) Collect() {
	cmd := bson.D{{Key: "replSetGetStatus", Value: "1"}}
	res := d.client.Database("admin").RunCommand(d.ctx, cmd)

	var m bson.M

	if err := res.Decode(&m); err != nil {
		if e, ok := err.(mongo.CommandError); ok {
			if e.Code == replicationNotYetInitialized || e.Code == replicationNotEnabled {
				return
			}
		}
		log.Printf("cannot get replSetGetStatus: %s\n", err)

		return
	}

	log.Println("** [ReplSet INFO]  replSetGetStatus result:")
	GetReplmembers(d.ctx, d.client)
	GetReplelection(d.ctx, d.client)

}

func newreplSetGetStatusCollector(c context.Context, cl *mongo.Client) *replSetGetStatusCollector {
	var a labelsGetter
	return &replSetGetStatusCollector{
		c, cl, true, a,
	}
}
func ReplSetStatus(c context.Context, client *mongo.Client) {
	replset := newreplSetGetStatusCollector(c, client)
	replset.Collect()
}

func GetReplmembers(ctx context.Context, client *mongo.Client) (string, error) {
	var replmembers proto.ReplicaSetStatus
	if err := client.Database("admin").RunCommand(ctx, primitive.M{"replSetGetStatus": 1}).Decode(&replmembers); err != nil {
		return "", err
	}
	//fmt.Println(replmembers.Members,replmembers.MyState)
	type myreplinfo map[string]interface{}

	type replInfostruct struct {
		ID string
		Name string
		StateStr string
		Uptime string
		Health float64
		Optime int
	}
	replallmemInfo := make([]myreplinfo, 0)
	for _, mem := range replmembers.Members {

		replInfo := make(myreplinfo)

		memUptime := int(mem.Uptime)
		memOptime := int(mem.OptimeDate)/1000    // 查出来的是ISODATE 转换为unix时间戳/1000,单位为s，多组的差值即为s


		memUptimeFormatDay, memUptimeFormatHour, memUptimeFormatMin, memUptimeFormatSec := dateformat.ResolveTime(memUptime)
		//fmt.Println(mem_uptime_format_day,mem_uptime_format_hour,mem_uptime_format_min,mem_uptime_format_sec)
		memUptimeFormat := strconv.Itoa(memUptimeFormatDay) + " Days " + strconv.Itoa(memUptimeFormatHour) + " Hours " + strconv.Itoa(memUptimeFormatMin) + " Mins " + strconv.Itoa(memUptimeFormatSec) + " Secs"
		//log.Printf("** [ReplSetMembers] members[%v]: %s, status: %v, uptime: %v\n", mem.ID, mem.Name, mem.Health, memUptimeFormat)
		// 是不是用bson 还是解析成map
		/* 如果是SECONDARY节点，计算延迟 */
		switch {
		case mem.StateStr == "PRIMARY":
			replInfo["StateStr"] = "PRIMARY"
		case mem.StateStr == "SECONDARY":
			replInfo["StateStr"] = "SECONDARY"
		case mem.StateStr == "ARBITER":
			replInfo["StateStr"] = "ARBITER"
		case mem.StateStr == "STARTUP":
			replInfo["StateStr"] = "STARTUP"
		case mem.StateStr == "STARTUP2":
			replInfo["StateStr"] = "STARTUP2"
		case mem.StateStr == "ROLLBACK":
			replInfo["StateStr"] = "ROLLBACK"
		case mem.StateStr == "UNKNOWN":
			replInfo["StateStr"] = "UNKNOWN"
		case mem.StateStr == "DOWN":
			replInfo["StateStr"] = "DOWN"
		case mem.StateStr == "UNKNOWN":
			replInfo["StateStr"] = "REMOVED"
		default:
			replInfo["StateStr"] = "The status of the replica set is not obtained"
		}

		replInfo["ID"] = mem.ID
		replInfo["Uptime"] = memUptimeFormat
		replInfo["Name"] = mem.Name
		replInfo["Health"] = mem.Health
		replInfo["Optime"] = memOptime


		replallmemInfo=append(replallmemInfo,replInfo)

	}

	//获取Primary节点的Optime时间戳，表示当前接节点的最新oplog时间
	priOptime:=0
	for _,v :=range replallmemInfo{
		if v["StateStr"]=="PRIMARY"{
			priOptime=v["Optime"].(int)
		}
	}
	var oplag string
	for _,v :=range replallmemInfo{
		if v["StateStr"]=="SECONDARY"{
			secOptime:=v["Optime"].(int)
			d,h,m,s:=dateformat.ResolveTime(priOptime-secOptime)
			oplag=fmt.Sprintf("%v Days %v Mins %v Hours %v Secs",d,h,m,s)
			log.Printf("** [ReplSetMembers] members[%v]: %s, role:SECONDARY, status: %v, uptime: %v, replsetLag: %v\n", v["ID"], v["Name"], v["Health"], v["Uptime"],oplag)
		}else {
			log.Printf("** [ReplSetMembers] members[%v]: %s, role:PRIMARY, status: %v, uptime: %v\n", v["ID"], v["Name"], v["Health"], v["Uptime"])
		}
	}
	return "", nil
}

func GetReplelection(ctx context.Context, client *mongo.Client) (elresion string, eltime string, err error) {
	var replmembers proto.ReplicaSetStatus
	var ts2unix int64
	if err := client.Database("admin").RunCommand(ctx, primitive.M{"replSetGetStatus": 1}).Decode(&replmembers); err != nil {
		return "", "0", err
	}

	eleResion := replmembers.ElectionCandidateMetrics.LastElectionReason
	eleTime := replmembers.ElectionCandidateMetrics.LastElectionDate
	ts2unix = int64(eleTime / 1000) // 这里直接取整即可 得到的就是unix时间戳
	tm := time.Unix(ts2unix, 0)
	eledate := tm.Format("2006-01-02 15:04:05") //得到的是ASIA/Shanghai的时间 与数据库中的比要多8小时

	nowts := time.Now().Unix()
	d, h, m, s := dateformat.ResolveTime(int(nowts - ts2unix))
	fromLastelection := fmt.Sprintf("%v Days %v Hours %v Mins %v Secs", d, h, m, s)
	log.Printf("** [ReplSetElection] ElectionResion: \"%s\",LastElectionDate: \"%s\",fromLastelection: \"%s\"", eleResion, eledate, fromLastelection) // 加上距离上次选举 有多长时间
	return eleResion, eledate, nil
}
