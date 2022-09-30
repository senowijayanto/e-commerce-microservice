package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MONGO_URI                  = "mongodb://admin:secret@mongodb_auth"
	MONGO_DATABASE             = "authDB"
	MONGO_POOL_MIN             = 10
	MONGO_POOL_MAX             = 100
	MONGO_MAX_IDLE_TIME_SECOND = 60
)

func NewMongoContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func NewMongoDatabase() *mongo.Database {
	ctx, cancel := NewMongoContext()
	defer cancel()

	option := options.Client().
		ApplyURI(MONGO_URI).
		SetMinPoolSize(uint64(MONGO_POOL_MIN)).
		SetMaxPoolSize(uint64(MONGO_POOL_MAX)).
		SetMaxConnIdleTime(time.Duration(MONGO_MAX_IDLE_TIME_SECOND) * time.Second)

	client, err := mongo.NewClient(option)
	if err != nil {
		log.Panic(err)
	}

	err = client.Connect(ctx)
	if err != nil {
		log.Panic(err)
	}

	database := client.Database(MONGO_DATABASE)
	return database
}
