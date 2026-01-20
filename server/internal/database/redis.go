package database

import (
	"context"
	"leaderboard/internal/config"
	"log"

	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisURL,
		Username: "default",
		Password: cfg.RedisPassword,
		DB:       0,
	})

	// Check connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("⚠️ Redis not reachable at %s. Falling back to Postgres only mode.", cfg.RedisURL)
		return nil
	}

	log.Println("✅ Successfully connected to Redis")
	return rdb
}

// func ExampleClient_connect_basic() {
// 	ctx := context.Background()

// 	rdb := redis.NewClient(&redis.Options{
// 		Addr:     "redis-.com:xxx",
// 		Username: "default",
// 		Password: "xxxxxxxxxxxxxxxxxx",
// 		DB:       0,
// 	})

// 	rdb.Set(ctx, "foo", "bar", 0)
// 	result, err := rdb.Get(ctx, "foo").Result()

// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println(result) // >>> bar

// }
