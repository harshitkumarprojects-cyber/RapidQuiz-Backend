package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/qiniu/qmgo"
)

var Client *qmgo.Client
var DB *qmgo.Database

func Connect() {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: uri})
	if err != nil {
		log.Fatal("Could not connect to MongoDB: ", err)
	}

	if err := client.Ping(int64(10)); err != nil {
		log.Fatal("MongoDB ping failed: ", err)
	}

	Client = client
	DB = client.Database(dbName)

	log.Println("Connected to MongoDB successfully")
}

func Collection(name string) *qmgo.Collection {
	return DB.Collection(name)
}

func Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Client.Close(ctx); err != nil {
		log.Println("Error disconnecting from MongoDB: ", err)
	}
}
