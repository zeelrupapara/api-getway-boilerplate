package manager

import (
	"sync"
	"testing"
	"greenlync-api-gateway/config"
	"greenlync-api-gateway/pkg/logger"
	"greenlync-api-gateway/pkg/shortuuid"

	"github.com/gofiber/websocket/v2"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {

	cfg := config.NewConfig()
	log, err := logger.NewLogger(cfg)
	require.NoError(t, err)
	hub := NewHub(log)

	wg := &sync.WaitGroup{}
	k := make(map[string]struct{})
	for i := 0; i < 100000; i++ {
		id := shortuuid.New()
		c := NewClient(&websocket.Conn{}, hub, id, int32(i), "123", k, 2)
		hub.Store(c)
		wg.Add(3)
		go func() {
			hub.Delete(id)
			wg.Done()
		}()
		// go func() {
		// 	hub.DeleteAll()
		// 	wg.Done()
		// }()
		go func() {
			hub.GetAll()
			wg.Done()
		}()
		go func() {
			hub.clientsCount()
			wg.Done()
		}()
	}
	wg.Wait()
}
