package mongodb

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	//	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// go.mongodb.org/mongo-driver/bson

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database

	log *zerolog.Logger

	cnclFunc context.CancelFunc
}

type (
	CollectionChats struct {
		ID       primitive.ObjectID       `bson:"_id"`
		Name     string                   `bson:"name"`
		AimId    string                   `bson:"aimId"`
		Messages []CollectionChatsMessage `bson:"messages"`
	}
	CollectionChatsMessage struct {
		MsgId  uint64    `bson:"msgId"`
		Time   time.Time `bson:"time"`
		Wid    string    `bson:"wid"`
		Sender string    `bson:"sender"`
		Text   string    `bson:"text"`
	}

	CollectionRAPIRequests struct {
		Method string                        `bson:"method"`
		ReqId  string                        `bson:"reqId"`
		Params *CollectionRAPIRequestsParams `bson:"params"`
	}
	CollectionRAPIRequestsParams struct {
		Sn           string `bson:"sn"`
		FromMsgId    uint64 `bson:"fromMsgId"`
		Count        int    `bson:"count"`
		PatchVersion string `bson:"patchVersion"`
	}

	// TODO - RESPONSE SAVE
)

/* test database credentials ( yes, i know; it's public data, ok? ):
host : 95.216.143.175:27017
user : icqdumper
pass : CHFiEEt9oQV05bMO6sudNRQ1 */

func NewMongoDriver(l *zerolog.Logger, mURI string) (mDriver *MongoDB, e error) {
	if mDriver.client, e = mongo.NewClient(options.Client().ApplyURI(mURI)); e != nil {
		return nil, e
	}

	mDriver.log = l
	mDriver.log.Info().Msg("MongoDB driver has been successfully inited")
	return mDriver, e
}

func (m *MongoDB) dbConnect() (e error) {
	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()

	m.cnclFunc = cncl
	defer func(m *MongoDB) { m.cnclFunc = nil }(m)

	if e = m.client.Connect(ctx); e != nil {
		m.log.Warn().Msg("Could not establish the connection with MongoDB")
		return e
	}
	if e = m.client.Ping(ctx, readpref.Primary()); e != nil {
		m.log.Warn().Msg("Could not ping MongoDB after successfully creating connection")
		return e
	}

	m.log.Info().Msg("Connection with MongoDB has been successfully established")
	return e
}

func (m *MongoDB) dbDisconnect() (e error) {
	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()

	if e = m.client.Disconnect(ctx); e != nil {
		m.log.Warn().Msg("Could not correctly close the MongoDB connection")
		return e
	}

	m.log.Info().Msg("Connection with MongoDB has been closed successfully")
	return e
}

func (m *MongoDB) dbInsertOne(collection string, data *interface{}) (e error) {
	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()

	if res, e := m.client.Database("icqdumper").Collection(collection).InsertOne(ctx, data); e == nil {
		m.log.Info().Str("inserted id", res.InsertedID.(primitive.ObjectID).Hex()).
			Msg("New record has been successfully writed")
	}

	return e
}

func (m *MongoDB) dbInsertMany(collection string, data *[]interface{}) (e error) {
	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()

	if res, e := m.client.Database("icqdumper").Collection(collection).InsertMany(ctx, *data); e == nil {
		for _, v := range res.InsertedIDs {
			m.log.Info().Str("inserted id", v.(primitive.ObjectID).Hex()).Msg("New record has been successfully writed")
		}
	}

	return e
}

func (m *MongoDB) dbUpdateOne(collection string, filter *interface{}, data *interface{}) (e error) {
	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()

	if res, e := m.client.Database("icqdumper").Collection(collection).UpdateOne(ctx, *filter, *data); e == nil {
		m.log.Info().Int64("matched", res.MatchedCount).Int64("modified", res.ModifiedCount).
			Msg("Some records in collection has been successfully updated")
	}

	return e
}

func (m *MongoDB) Construct() error { return m.dbConnect() }
func (m *MongoDB) Destruct() error {
	if m.cnclFunc != nil {
		m.cnclFunc() // race condition !!!
	}

	return m.dbDisconnect()
}
func (m *MongoDB) InsertOne(collection string, data *interface{}) (e error) {
	return m.dbInsertOne(collection, data)
}
func (m *MongoDB) InsertMany(collection string, data *[]interface{}) (e error) {
	return m.dbInsertMany(collection, data)
}
func (m *MongoDB) UpdateOne(collection string, filter *interface{}, data interface{}) (e error) {
	return e
}
