package oauth2

import (
	"context"
	"sync"
	"testing"
	"greenlync-api-gateway/config"
	"greenlync-api-gateway/pkg/cache"
	"greenlync-api-gateway/pkg/db"
	"greenlync-api-gateway/pkg/logger"
	"greenlync-api-gateway/pkg/redis"
	"greenlync-api-gateway/pkg/shortuuid"

	"github.com/stretchr/testify/require"
)

func TestSessions(t *testing.T) {
	cfg := config.NewConfig()
	log, err := logger.NewLogger(cfg)
	require.NoError(t, err)

	redis, err := redis.NewRedisClient(cfg)
	require.NoError(t, err)
	cache := cache.NewCache(redis)

	db, err := db.NewMysqDB(cfg)
	require.NoError(t, err)

	oauth := NewOAuth2(cache, db.DB, cfg, log)

	wg := &sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		id := shortuuid.New()
		cg := &Config{ClientId: 1, ClientSecretId: "", SessionId: id}
		oauth.NewActiveSession(cg)
		wg.Add(4)
		go func() {
			oauth.DeleteSessionId(context.Background(), id)
			wg.Done()
		}()
		go func() {
			oauth.NewActivity(id)
			wg.Done()
		}()
		go func() {
			oauth.GetActiveSessionById(cg.Id)
			wg.Done()
		}()
		go func() {
			oauth.GetActiveSessions()
			wg.Done()
		}()
	}
	wg.Wait()
}
