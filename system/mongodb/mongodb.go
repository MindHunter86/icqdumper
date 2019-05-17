package mongodb

import (
	"context"
	"time"

	"github.com/rs/zerolog"
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
		Name     string
		AimId    string
		Messages []CollectionChatsMessage
	}
	CollectionChatsMessage struct {
		MsgId  uint64
		Time   uint64
		Wid    string
		Sender string
		Text   string
	}
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

func (m *MongoDB) dbInsertOne(collection string, data interface{}) (e error) {
	ctx, cncl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cncl()

	var res *mongo.InsertOneResult
	if res, e = m.client.Database("icqdumper").Collection(collection).InsertOne(ctx, &data); e != nil {
		return e
	}

	m.log.Info().Interface("inserted id", res.InsertedID).Msg("New record has been successfully writed")

	return e
}

func (m *MongoDB) Construct() error { return m.dbConnect() }
func (m *MongoDB) Destruct() error {
	if m.cnclFunc != nil {
		m.cnclFunc() // race condition !!!
	}

	return m.dbDisconnect()
}
