// By Emran A. Hamdan, Lead Architect
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
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
func NewCache(redis *redis.Client) *Cache {
	return &Cache{redis: redis}
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
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// // Get Redis `GET key` command. It returns redis.Nil error when key does not exist.
// func (e *Cache) GetAll(ctx context.Context, key string) (string, error) {
// 	result, err := e.Redis.Sc(ctx, key).Result()
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(result), nil
// 	return "", nil
// }

// delete key if exists
func (e *Cache) Delete(ctx context.Context, key string) error {
	return e.redis.Del(ctx, key).Err()
}

func (e *Cache) FlushAll(ctx context.Context) error {
	return e.redis.FlushAll(ctx).Err()
}

// Ping checks if Redis connection is alive
func (e *Cache) Ping(ctx context.Context) error {
	return e.redis.Ping(ctx).Err()
}

// GetRedisClient returns the underlying Redis client for advanced operations
func (e *Cache) GetRedisClient() *redis.Client {
	return e.redis
}
