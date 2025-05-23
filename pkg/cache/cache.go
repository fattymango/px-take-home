package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/fattymango/px-take-home/config"
	"github.com/redis/go-redis/v9"
)

var (
	KeySessionsMap = "sessions_map"

	SessionsKey = func(sessionId string) string { return fmt.Sprint("sessions_", sessionId) }
	TokensKey   = func(token string) string { return fmt.Sprint("tokens_", token) }
	RefreshKey  = func(refreshToken string) string { return fmt.Sprint("refresh_tokens_", refreshToken) }
)

type Cache struct {
	redis *redis.Client
}

// NewCache return new cache instant constructor using redisClient
func NewCache(c *config.Config) (*Cache, error) {
	r, err := NewRedisClient(c)
	if err != nil {
		return nil, err
	}
	return &Cache{redis: r}, nil
}

// Set Redis `SET key value [expiration]` command.
// Use expiration for `SETEX`-like behavior.
//
// Zero expiration means the key has no expiration time.
// KeepTTL is a Redis KEEPTTL option to keep existing TTL, it requires your redis-server version >= 6.0,
// otherwise you will receive an error: (error) ERR syntax error.
func (e *Cache) Set(ctx context.Context, key string, value []byte, expiration int) error {
	return e.redis.Set(ctx, key, string(value), time.Duration(expiration)*time.Second).Err()
}

// Get Redis `GET key` command. It returns redis.Nil error when key does not exist.
func (e *Cache) Get(ctx context.Context, key string) (string, error) {
	result, err := e.redis.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return "", err
	}

	return string(result), nil
}

// delete key if exists
func (e *Cache) Delete(ctx context.Context, key string) error {
	return e.redis.Del(ctx, key).Err()
}

func (e *Cache) FlushAll(ctx context.Context) error {
	return e.redis.FlushAll(ctx).Err()
}

func (e *Cache) IsRedisNil(err error) bool {
	return err == redis.Nil
}
