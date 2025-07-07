// Developer: Saif Hamdan
// Date: 18/7/2023

package redis

import (
	"testing"
	"greenlync-api-gateway/config"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func TestNewRedis(t *testing.T) {

	testCases := []struct {
		RedisAddr     string
		RedisPassword string
		CheckResponse func(t *testing.T, red *redis.Client, err error)
	}{
		{
			RedisAddr:     "0.0.0.0:6379",
			RedisPassword: "null",
			CheckResponse: func(t *testing.T, red *redis.Client, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, red)
			},
		},
		{
			RedisAddr:     "0.0.0.0:7777",
			RedisPassword: "null",
			CheckResponse: func(t *testing.T, red *redis.Client, err error) {
				require.Error(t, err)
				require.Nil(t, red)
			},
		},
	}

	for i := range testCases {
		red, err := NewRedisClient(&config.Config{
			Redis: config.Redis{
				RedisAddr:     testCases[i].RedisAddr,
				RedisPassword: testCases[i].RedisPassword,
			},
		})
		testCases[i].CheckResponse(t, red, err)
	}
}
