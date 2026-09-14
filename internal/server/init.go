package server

import (
	"context"
	"fmt"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

var limiter *redis_rate.Limiter

func InitRedis(ctx context.Context) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDISPASSWORD"),
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	limiter = redis_rate.NewLimiter(rdb)
	return rdb, nil
}

func InitClickhouse() (driver.Conn, error) {
	newConn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"clickhouse:8123"},
		Auth: clickhouse.Auth{
			Database: "logs",
			Username: "default",
			Password: "",
		},
		Protocol: clickhouse.HTTP,
	})
	if err != nil {
		return nil, err
	}
	fmt.Print("Currently initialising clickhouse")
	return newConn, nil
}
