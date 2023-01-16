package rdb

import (
	"context"
	"log"

	"github.com/go-redis/redis/v9"
)

var ctx = context.Background()

func Init() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	log.Println("Connected to redis...")
	return rdb
}
