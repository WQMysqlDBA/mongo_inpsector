package AccessMongoDB

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"mongostatus/Output"
	"strconv"
	"strings"
)

type indexstatsCollector struct {
	ctx             context.Context
	client          *mongo.Client
	collections     []string
	discoveringMode bool
}

func (d *indexstatsCollector) Collect(fl, fi string) {

	collections := d.collections
	if d.discoveringMode {
		namespaces, err := listAllCollections(d.ctx, d.client, d.collections)
		if err != nil {
			msg := fmt.Sprintf("cannot auto discover databases and collections")
			Output.DoResult(msg, fl)
			return
		}
		collections = Map2Slice(namespaces) //[ycsb.test]

	}

	for _, dbCollection := range collections {

		parts := strings.Split(dbCollection, ".")
		if len(parts) < 2 { //nolint:gomnd
			continue
		}
		database := parts[0]
		collection := strings.Join(parts[1:], ".")

		/* 过滤掉system.profile这个集合 */
		if collection == "system.profile" {
			continue
		}

		aggregation := bson.D{
			{Key: "$indexStats", Value: bson.M{}},
		}

		cursor, err := d.client.Database(database).Collection(collection).Aggregate(d.ctx, mongo.Pipeline{aggregation})
		if err != nil {
			msg := fmt.Sprintf("cannot get $indexStats cursor for collection %s.%s: %s", database, collection, err)
			Output.DoResult(msg, fl)
			continue
		}

		//fmt.Printf("%T\n", cursor) ----> *mongo.Cursor

		// Iterate the cursor and print out each document until the cursor is
		// exhausted or there is an error getting the next document.

		i := 0
		var indexslice []string
		for cursor.Next(context.TODO()) {
			// A new result variable should be declared for each document.
			var result bson.M
			if err := cursor.Decode(&result); err != nil {
				log.Fatal(err)
			}

			/* 这样得不到 。。。。 */
			//m, ok := result["key"].(bson.M)
			//if ok {
			//	for k, v := range m {
			//		fmt.Println(k, v)
			//	}
			//}

			indexslice = append(indexslice, Strval(RtuStringFromBsonM(result)))
			i++
		}
		if i != 0 {
			// has indexes
			msg := fmt.Sprintf(" ** [IndexInfo] collection: %s has %v index,is %s", dbCollection, i, indexslice)
			Output.DoResult(msg, fl)
			variable := fmt.Sprintf(" ** [VARIABLES]: VAR_INDEX:%s%s%v%s%s", dbCollection, _splitflag, i, _splitflag, indexslice)
			Output.DoResult(variable, fl)

			Output.Writeins(fmt.Sprintf("**[IndexInfo] collection: %s has %v index**\n", dbCollection, i), fi)
			Output.Writeins("```txt", fi)
			for k, v := range indexslice {
				m := "index[" + strconv.Itoa(k) + "]: " + v
				Output.Writeins(m, fi)
			}
			Output.Writeins("```", fi)

		} else {
			// No indexes
			msg := fmt.Sprintf(" ** [IndexInfo] collection: %s has %v index", dbCollection, i)
			Output.DoResult(msg, fl)
			variable := fmt.Sprintf("VAR_INDEX:%s#%v", dbCollection, i)
			Output.DoResult(variable, fl)
		}

		if err := cursor.Err();
			err != nil {
			log.Fatal(err)
		}

		var stats []bson.M
		if err = cursor.All(d.ctx, &stats); err != nil {
			msg := fmt.Sprintf("cannot get $indexStats for collection %s.%s: %s", database, collection, err)
			Output.DoResult(msg, fl)
			continue
		}
		_ = cursor.Close(context.TODO())
	}
}

func newindexstatsCollector(c context.Context, cl *mongo.Client, dbname string) *indexstatsCollector {
	filter := bson.D{} // default : empty
	collection, _ := cl.Database(dbname).ListCollectionNames(c, filter)
	return &indexstatsCollector{
		c, cl, collection, true,
	}
}

func DBIndex(c context.Context, client *mongo.Client, dbname string, f, f1 string) {
	indexstats := newindexstatsCollector(c, client, dbname)
	indexstats.Collect(f, f1)
}

func RtuStringFromBsonM(b bson.M) (idx interface{}) {
	/* mongodb的索引格式如下：
	{
			"v" : 2,
			"key" : {
				"_id" : 1
			},
			"name" : "_id_"
		},
		{
			"v" : 2,
			"key" : {
				"userId" : 1
			},
			"name" : "userId_1"
		}
	bson.M是一个无序的map
	格式如下：
	map[
	    accesses:map[
	        ops:500421
	        since:1638860881822
	        ]

	    host:mongo-20:27017

	    key:map[
	        _id:1
	        ]

	    name:_id_

	    spec:map[
	        key:map[
	            _id:1
	        ]
	        name:_id_
	        v:2
	    ]
	]
	因此，只获取最大的一个map的里面的子map spec即可得到索引的信息
	key 为_id
	name 为index的名称这个可以不用太关注
	v 表示索引的版本 这个是默认的 也可以不管
	*/

	key := b["spec"]
	return key
}

func Strval(value interface{}) string {
	var key string
	if value == nil {
		return key
	}
	// 做断言，返回
	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}
