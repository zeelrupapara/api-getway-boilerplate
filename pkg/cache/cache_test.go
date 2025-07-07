package cache

import (
	"context"
	"testing"
	"greenlync-api-gateway/config"
	"greenlync-api-gateway/pkg/redis"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheSetAndGet(t *testing.T) {
	redis, err := redis.NewRedisClient(&config.Config{
		Redis: config.Redis{
			RedisAddr: "0.0.0.0:6379",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, redis)

	cache := NewCache(redis)
	require.NotNil(t, cache)

	t.Run("setAndGet", func(t *testing.T) {
		ctx := context.Background()

		// Test Set method
		err = cache.Set(ctx, "key1", []byte("value1"), 10)
		require.NoError(t, err)

		// Test Get method
		value, err := cache.Get(ctx, "key1")
		require.NoError(t, err)
		require.Equal(t, "value1", value)

		// Test Get method for non-existent key
		value, err = cache.Get(ctx, "nonexistent")
		require.Error(t, err)
		require.Equal(t, "", value)

	})

	t.Run("delete", func(t *testing.T) {
		ctx := context.Background()

		// Test Delete method
		err = cache.Delete(ctx, "key1")
		assert.NoError(t, err)

		// Verify the key is deleted
		value, err := cache.Get(ctx, "key1")
		assert.Error(t, err)
		assert.Equal(t, "", value)
	})
}
