package config

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)


func GetRedis() *redis.Client {
	redisOnce.Do(func() {
		if url := os.Getenv("REDIS_URL"); url != "" {
			opt, err := redis.ParseURL(url)
			if err != nil {
				log.Fatalf("failed parse REDIS_URL: %v", err)
			}
			redisClient = redis.NewClient(opt)
		} else {
			addr := os.Getenv("REDIS_ADDR")
			if addr == "" {
				addr = "localhost:6379"
			}
			pass := os.Getenv("REDIS_PASSWORD")
			db := 0
			if v := os.Getenv("REDIS_DB"); v != "" {
				if parsed, err := strconv.Atoi(v); err == nil {
					db = parsed
				}
			}
			redisClient = redis.NewClient(&redis.Options{
				Addr:         addr,
				Password:     pass,
				DB:           db,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
			})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Fatalf("failed to connect to redis: %v", err)
		}
		log.Println("connected to redis")
	})

	return redisClient
}

func CloseRedis() error {
	if redisClient == nil {
		return nil
	}
	return redisClient.Close()
}
