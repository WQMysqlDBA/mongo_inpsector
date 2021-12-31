package proto

/*
	"hosts" : [
		"mongo-20:27017",
		"172.18.10.21:27017",
		"172.18.10.22:27017"
	],
	"setName" : "rs1",

*/
type MasterDoc struct {
	SetName interface{} `bson:"setName"`  //
	Hosts   interface{} `bson:"hosts"`
	Msg     string      `bson:"msg"`
}
