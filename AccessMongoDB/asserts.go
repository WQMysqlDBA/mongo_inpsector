package AccessMongoDB

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"mongostatus/proto"
	"mongostatus/util"
)

//db.RunCommandDBstats({getDiagnosticData:1}).data.serverStatus.asserts
//db.RunCommandDBstats({getDiagnosticData:1}).data.serverStatus.asserts.connections
//db.RunCommandDBstats({getDiagnosticData:1}).data."local.oplog.rs.stats"  ======
//db.RunCommandDBstats({getDiagnosticData:1}).data.systemMetrics
// 以下三个用replSetGetStatus
//db.RunCommandDBstats({getDiagnosticData:1}).data.replSetGetStatus.majorityVoteCount
//db.RunCommandDBstats({getDiagnosticData:1}).data.replSetGetStatus.writeMajorityCount
//db.RunCommandDBstats({getDiagnosticData:1}).data.replSetGetStatus.votingMembersCount
type DiagnosticDatastr struct {
	Data Data `bson:"data"`
}

type Data struct {
	ServiceStatus proto.ServerStatus `bson:"serverStatus"`
}

type Assertmod struct {
	ctx    context.Context
	client *mongo.Client
}

func (d *Assertmod) Collect() {
	//cmd := bson.D{{Key: "getDiagnosticData", Value: "1"}}
	//res := d.client.Database("admin").RunCommand(d.ctx, cmd)
	//var m bson.M
	service, err := util.GetServerStatus(d.ctx, d.client)
	if err != nil {
		log.Println(err.Error())
	}
	fmt.Println(service.Asserts)

}

func newAssert(c context.Context, cl *mongo.Client) *Assertmod {
	return &Assertmod{c, cl}
}
func GetAsserts(ctx context.Context, client *mongo.Client) {
	assert := newAssert(ctx, client)
	assert.Collect()
}
