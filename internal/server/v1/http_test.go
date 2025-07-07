package v1

import (
	"context"
	"fmt"
	"greenlync-api-gateway/config"
	"greenlync-api-gateway/internal/middleware"
	"greenlync-api-gateway/pkg/authz"
	"greenlync-api-gateway/pkg/cache"
	"greenlync-api-gateway/pkg/db"
	app "greenlync-api-gateway/pkg/http"
	"greenlync-api-gateway/pkg/logger"
	"greenlync-api-gateway/pkg/manager"
	"greenlync-api-gateway/pkg/nats"
	"greenlync-api-gateway/pkg/oauth2"
	"greenlync-api-gateway/pkg/redis"
	"greenlync-api-gateway/pkg/smtp"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"
)

// for testing only
func NewTestServer(t *testing.T) *HttpServer {

	// config instant
	cfg := config.NewConfig()

	// pass to logger handler instant
	log, _ := logger.NewLogger(cfg)

	// Languages Translate
	// local, err := i18n.New(cfg, "en-US", "el-GR", "zh-CN")
	// if err != nil {
	// 	log.Logger.Fatalf("failed to init i18n package %v", err)
	// }

	// Start log
	log.Logger.Info("Logging started for testing:")

	// connect to redis
	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		log.Logger.Fatalf("Error connectting to redis service at %v", cfg.Redis.RedisAddr)
	}

	if redisClient != nil {
		log.Logger.Infof("Connected to redis %s", cfg.Redis.RedisAddr)
	}

	// make our cache wrapper from Redis
	cacheClient := cache.NewCache(redisClient)
	if log == nil {
		fmt.Println("Could not start a cache session")
		panic(0)
	}

	// connect to mysql using gorm and grap a session
	dbSess, err := db.NewMysqDB(cfg)
	if err != nil {
		fmt.Printf("We have a problem connecting to database %v", err)
		panic(0)
	}

	// This is the best time migrate in case you change the schema
	err = dbSess.Migrate()
	if err != nil {
		fmt.Printf("We have a problem mirgrating tables %v", err)
		panic(0)
	}

	// authorization
	authz, err := authz.NewAuthz(dbSess.DB)
	if err != nil {
		fmt.Printf("We have a problem creating authorization %v", err)
		panic(0)
	}

	//validetor
	validate := validator.New()

	// Nats
	nats, err := nats.NewNatClient(cfg)
	if err != nil {
		log.Logger.Fatalf("Error Connecting to Nats: %v", err)
	}

	// go-corn
	corn := gocron.NewScheduler(time.UTC)

	smtp, err := smtp.NewSmtpClient(cfg, log, dbSess, corn)
	if err != nil {
		log.Logger.Error("Error Connecting to SMTP Client: %v", err)
	}

	// fiber instence
	app := app.NewApp(log)

	// Websocket Hub
	newHub := manager.NewHub(log)
	manager.SetMaxWebsocketConnections()
	// OAuth2
	oauth2 := oauth2.NewOAuth2(cacheClient, dbSess.DB, cfg, log)
	// middleware
	middleware := middleware.NewMiddleware(app, dbSess.DB, authz, oauth2, log, nats)

	// go-corn
	cron := gocron.NewScheduler(time.UTC)
	cron.StartAsync()

	// v1 HTTP
	newHttp := NewHTTP(app, dbSess.DB, log, cacheClient, nats, authz, oauth2, newHub, middleware, smtp, cfg, validate, cron)

	newHttp.RegisterV1()

	return newHttp
}

func (server *HttpServer) GetTestToken(t *testing.T) *oauth2.Config {
	cfg := &oauth2.Config{
		// ClientName:     user.Name,
		Scope:          "admin",
		ClientId:       1,
		ClientSecretId: "",
		IpAddress:      "0.0.0.0",
	}

	ctx := context.Background()
	_, err := server.OAuth2.PasswordCredentialsToken(ctx, cfg)
	require.NoError(t, err)

	return cfg
}

func ResponseCheckerOk(t *testing.T, resp *http.Response, statusCode int) []byte {
	require.Equal(t, statusCode, resp.StatusCode)
	require.NotEmpty(t, resp.Body)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	body := &app.HttpResponse{}
	err = json.Unmarshal(b, body)
	require.NoError(t, err)

	require.Equal(t, statusCode, body.Code)
	require.Equal(t, true, body.Success)
	require.Empty(t, body.Error)

	b, err = json.Marshal(body.Data)
	require.NoError(t, err)

	return b
}

func ResponseCheckerBad(t *testing.T, resp *http.Response, statusCode int) {
	require.Equal(t, statusCode, resp.StatusCode)
	require.NotEmpty(t, resp.Body)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	body := &app.HttpResponse{}
	err = json.Unmarshal(b, body)
	require.NoError(t, err)

	require.Equal(t, statusCode, body.Code)
	require.Equal(t, false, body.Success)
	require.NotEmpty(t, body.Error)
	require.Empty(t, body.Data)
}

func ResponseCheckerNoContent(t *testing.T, resp *http.Response) {
	require.Equal(t, 204, resp.StatusCode)
	require.Empty(t, resp.Body)
}
