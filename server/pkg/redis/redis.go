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

	opt, _ := redis.ParseURL(redisURL)
	rdb := redis.NewClient(opt)

	_, err := rdb.Ping(context.TODO()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis client created and connected successfully...")

	return rdb
}
