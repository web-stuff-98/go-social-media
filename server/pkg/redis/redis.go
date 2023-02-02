package rdb

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v9"
)

func Init() *redis.Client {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	var rdb *redis.Client

	opt, _ := redis.ParseURL(redisURL)
	if os.Getenv("PRODUCTION") == "true" {
		rdb = redis.NewClient(opt)
	} else {
		rdb = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
	}

	_, err := rdb.Ping(context.TODO()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis client created and connected successfully...")

	return rdb
}
