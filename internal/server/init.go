package server

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

var limiter *redis_rate.Limiter

func InitRedis(ctx context.Context) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDISPASSWORD"),
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	limiter = redis_rate.NewLimiter(rdb)
	return rdb, nil
}
