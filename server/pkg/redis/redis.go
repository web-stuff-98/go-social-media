package rdb

import (
	"log"

	"github.com/go-redis/redis/v9"
)

func Init() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	log.Println("Redis client created...")
	return rdb
}
