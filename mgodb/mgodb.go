package mgodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"time"
)

type mgo struct {
	database   string
	collection string
}

func NewMgo(database, collection string) *mgo {

	return &mgo{
		database,
		collection,
	}
}

// 查询单个
func (m *mgo) FindOne(ctx context.Context, client *mongo.Client, key string, value interface{}) *mongo.SingleResult {

	collection, _ := client.Database(m.database).Collection(m.collection).Clone()
	//collection.
	filter := bson.D{{key, value}}
	singleResult := collection.FindOne(context.Background(), filter)
	return singleResult
}

//插入单个
func (m *mgo) InsertOne(ctx context.Context, client *mongo.Client, value interface{}) *mongo.InsertOneResult {

	collection := client.Database(m.database).Collection(m.collection)
	insertResult, err := collection.InsertOne(context.Background(), value)
	if err != nil {
		fmt.Println(err)
	}
	return insertResult
}

//查询集合里有多少数据
func (m *mgo) CollectionCount(ctx context.Context, client *mongo.Client) (string, int64) {

	collection := client.Database(m.database).Collection(m.collection)
	name := collection.Name()
	size, _ := collection.EstimatedDocumentCount(context.Background())
	return name, size
}

//按选项查询集合 Skip 跳过 Limit 读取数量 sort 1 ，-1 . 1 为最初时间读取 ， -1 为最新时间读取
func (m *mgo) CollectionDocuments(ctx context.Context, client *mongo.Client, Skip, Limit int64, sort int,sortkey string) *mongo.Cursor {
	collection := client.Database(m.database).Collection(m.collection)
	//fmt.Println(m.database,m.collection)
	SORT := bson.D{{sortkey, sort}} //filter := bson.D{{key,value}}
	filter := bson.D{{}}
	findOptions := options.Find().SetSort(SORT).SetLimit(Limit).SetSkip(Skip)
	//findOptions.SetLimit(i)
	temp, _ := collection.Find(context.Background(), filter, findOptions)
	return temp
}

//获取集合创建时间和编号
func (m *mgo) ParsingId(result string) (time.Time, uint64) {
	temp1 := result[:8]
	timestamp, _ := strconv.ParseInt(temp1, 16, 64)
	dateTime := time.Unix(timestamp, 0) //这是截获情报时间 时间格式 2019-04-24 09:23:39 +0800 CST
	temp2 := result[18:]
	count, _ := strconv.ParseUint(temp2, 16, 64) //截获情报的编号
	return dateTime, count
}

//删除文章和查询文章
func (m *mgo) DeleteAndFind(ctx context.Context, client *mongo.Client, key string, value interface{}) (int64, *mongo.SingleResult) {

	collection := client.Database(m.database).Collection(m.collection)
	filter := bson.D{{key, value}}
	singleResult := collection.FindOne(context.Background(), filter)
	DeleteResult, err := collection.DeleteOne(context.Background(), filter, nil)
	if err != nil {
		fmt.Println("删除时出现错误，你删不掉的~")
	}
	return DeleteResult.DeletedCount, singleResult
}

//删除文章
func (m *mgo) Delete(ctx context.Context, client *mongo.Client, key string, value interface{}) int64 {

	collection := client.Database(m.database).Collection(m.collection)
	filter := bson.D{{key, value}}
	count, err := collection.DeleteOne(context.Background(), filter, nil)
	if err != nil {
		fmt.Println(err)
	}
	return count.DeletedCount

}

//删除多个
func (m *mgo) DeleteMany(ctx context.Context, client *mongo.Client, key string, value interface{}) int64 {

	collection := client.Database(m.database).Collection(m.collection)
	filter := bson.D{{key, value}}

	count, err := collection.DeleteMany(context.Background(), filter)
	if err != nil {
		fmt.Println(err)
	}
	return count.DeletedCount
}
