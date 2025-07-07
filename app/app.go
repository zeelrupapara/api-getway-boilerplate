// Developer 	: zeelrupapara@gmail.com
// Last Update  : Cannabis boilerplate conversion
// Update reason : Convert VFX trading platform to cannabis compliance system
package app

import (
	"context"
	"fmt"
	"greenlync-api-gateway/config"
	"greenlync-api-gateway/internal/server"
	"greenlync-api-gateway/pkg/authz"
	"greenlync-api-gateway/pkg/cache"
	"greenlync-api-gateway/pkg/db"
	"greenlync-api-gateway/pkg/i18n"
	// Removed influxdb for minimal boilerplate
	"greenlync-api-gateway/pkg/logger"
	"greenlync-api-gateway/pkg/monitor"
	"greenlync-api-gateway/pkg/nats"
	"greenlync-api-gateway/pkg/redis"
	"greenlync-api-gateway/pkg/smtp"
	"greenlync-api-gateway/utils"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/go-playground/validator/v10"
)

var (
	// this name is one time only
	service = "greenlync-api-gateway"

	// this change as per git -tag -v everytime this will go into testing
	// v1.0.0 Major.Minor.Batch or bug

	version = "v1.0.0"
)

func Start() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// config instant
	cfg := config.NewConfig()

	// pass to logger handler instant
	log, _ := logger.NewLogger(cfg)

	// Languages Translate
	local, err := i18n.New(cfg, "en-US", "el-GR", "zh-CN")
	if err != nil {
		log.Logger.Fatalf("failed to init i18n package %v", err)
	}

	// Start logging
	log.Logger.Info("Logging started for service: ", service+"@"+version)

	// Init Monitor
	monitor.InitMonitor()

	// Check for the license and validate
	// lm, err := license.NewLicenseMgr(cfg.License.COMPANY_ID, cfg.License.COMPANY_NAME)
	// if err != nil {
	// 	log.Logger.Fatalf("failed to init license manager %v", err)
	// }

	// // Validate the license
	// err = lm.Validate()
	// if err != nil {
	// 	log.Logger.Fatalf("failed to validate license %v", err)
	// }

	// connect to redis
	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		log.Logger.Fatalf("Error connectting to redis service at %v", cfg.Redis.RedisAddr)
	}

	if redisClient != nil {
		log.Logger.Infof("Connected to redis %s", cfg.Redis.RedisAddr)
		monitor.ChangeStatus(monitor.Health_Redis, monitor.HealthStatus_running)
	}

	// make our cache wrapper from Redis
	cacheClient := cache.NewCache(redisClient)

	// connect to mysql using gorm and grap a session
	dbSess, err := db.NewMysqDB(cfg)
	if err != nil {
		monitor.ChangeStatus(monitor.Health_Database, monitor.HealthStatus_error)
		fmt.Printf("We have a problem connecting to database %v", err)
		panic(0)
	}
	monitor.ChangeStatus(monitor.Health_Database, monitor.HealthStatus_running)

	// This is the best time migrate in case you change the schema
	err = dbSess.Migrate()
	if err != nil {
		fmt.Printf("We have a problem mirgrating tables %v", err)
		panic(0)
	}

	// validate db data
	err = dbSess.ValidateDBData()
	if err != nil {
		log.Logger.Warnf("Database validation warning: %v", err)
		log.Logger.Info("Please run database seeding to create initial users and data")
	}

	// convert auto increment Id to a custom id
	err = dbSess.AlterTableIds()
	if err != nil {
		fmt.Printf("We have a problem converting auto increment Id to a custom id %v", err)
		panic(0)
	}

	// Cannabis compliance setup can be added here in the future
	// Example: Cannabis license validation, compliance checks, etc.

	// Nats
	nats, err := nats.NewNatClient(cfg)
	if err != nil {
		monitor.ChangeStatus(monitor.Health_NATS, monitor.HealthStatus_error)
		log.Logger.Fatalf("Error Connecting to Nats: %v", err)
	}
	monitor.ChangeStatus(monitor.Health_NATS, monitor.HealthStatus_running)

	// authorization
	authz, err := authz.NewAuthz(dbSess.DB)
	if err != nil {
		fmt.Printf("We have a problem creating authorization %v", err)
		panic(0)
	}

	// validetor
	validate := validator.New()

	// register custom validation functions
	validate.RegisterValidation("username", utils.UsernameValidation)

	// go-corn
	cron := gocron.NewScheduler(time.UTC)
	cron.StartAsync()

	// Removed InfluxDB for minimal event-driven boilerplate

	// SMTP
	smtp, err := smtp.NewSmtpClient(cfg, log, dbSess, cron)
	if err != nil {
		monitor.ChangeStatus(monitor.Health_SMTP, monitor.HealthStatus_error)
		log.Logger.Error("Error Connecting to SMTP: %v", err)
	}
	monitor.ChangeStatus(monitor.Health_SMTP, monitor.HealthStatus_running)

	// http API server based on fiber
	server := server.NewServer(local, log, cacheClient, dbSess.DB, authz, nats, validate, cfg, smtp, cron)

	// Register all APP APIs
	// this should be before listening
	server.Register()

	// start http server
	go func() {
		err := server.App.Listen(cfg.HTTP.Host + cfg.HTTP.Port)
		if err != nil {
			log.Logger.Fatalf("Error trying to listenning on port %s: %v", cfg.HTTP.Port, err)
		}
	}()

	// when painc receover
	if err := recover(); err != nil {
		log.Logger.Fatalf("some panic ...:", err)
	}

	// we need nice way to exit will use os package notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case v := <-quit:
		fmt.Printf("signal.Notify CTRL+C: %v", v)
	case done := <-ctx.Done():
		fmt.Printf("ctx.Done: %v", done)
	}

	// graceful shutdown completed
	log.Logger.Info("Server shutdown completed")
}
