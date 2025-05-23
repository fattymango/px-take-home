package cache

import (
	"context"
	"fmt"

	"github.com/fattymango/px-take-home/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		// MinIdleConns: cfg.Redis.MinIdleConn,
		// PoolSize:     cfg.Redis.PoolSize,
		// PoolTimeout:  time.Duration(cfg.Redis.PoolTimeout) * time.Second,
		//Password:     cfg.Redis.RedisPassword, // no password set
		//DB:           cfg.Redis.DB,            // use default DB
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
