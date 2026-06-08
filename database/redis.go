package database

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis() {
	log.Println("REDIS_ADDR:", os.Getenv("REDIS_ADDR"))
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx := context.Background()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Could not connect to Redis: ", err)
	}

	log.Println("Connected to Redis successfully")
}