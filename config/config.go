// Developer 	: zeelrupapara@gmail.com
// Last Update  : Cannabis boilerplate conversion
// Update reason : Convert VFX config to cannabis compliance system

package config

// Config will use .ENV for docker-compose and load into config
import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Env vars gose here so we don't change names by mistake
const (
	BASE_URL                    = "BASE_URL"
	LOGS_FILE                   = "LOG_FILE"
	HTTP_HOST                   = "HTTP_HOST"
	HTTP_PORT                   = "HTTP_PORT"
	OAUTH_TOKEN_EXPIRES_IN      = "OAUTH_TOKEN_EXPIRES_IN"
	OAUTH_LONG_TOKEN_EXPIRES_IN = "OAUTH_LONG_TOKEN_EXPIRES_IN"
	APP_REPORTS                 = "APP_REPORTS"
	GRPC_HOST                   = "GRPC_HOST"
	GRPC_PORT                   = "GRPC_PORT"
	REDIS_URL                   = "REDIS_URL"
	REDIS_PASSWORD              = "REDIS_PASSWORD"
	MYSQL_HOST                  = "MYSQL_HOST"
	MYSQL_PORT                  = "MYSQL_PORT"
	MYSQL_USER                  = "MYSQL_USER"
	MYSQL_PASSWORD              = "MYSQL_PASSWORD"
	MYSQL_DB                    = "MYSQL_DB"
	MYSQL_USE_SSL               = "MYSQL_USE_SSL"
	MYSQL_CACERT_PATH           = "MYSQL_CACERT_PATH"
	NATS_HOST                   = "NATS_HOST"
	NATS_PORT                   = "NATS_PORT"
	SMTP_HOST                   = "SMTP_HOST"
	SMTP_PORT                   = "SMTP_PORT"
	SMTP_FROM                   = "SMTP_FROM"
	SMTP_PASSWORD               = "SMTP_PASSWORD"
	SMTP_LOGIN                  = "SMTP_LOGIN"
)

// Config blueprint microservice
type Config struct {
	Setting       Setting
	GRPC          GRPC
	Logger        Logger
	Redis         Redis
	MySQL         MySQL
	HTTP          Http
	Nats          Nats
	Smtp          SMTP
}

type Setting struct {
	Version   string
	LocalPath string
}

// Logger config
type Logger struct {
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	Level             string
	LogFile           string
}

// Redis config
type Redis struct {
	RedisAddr      string
	RedisPassword  string
	RedisDB        string
	RedisDefaultDB string
	MinIdleConn    int
	PoolSize       int
	PoolTimeout    int
	DB             int
}

// Mongo
type Mongo struct {
	URI         string
	PoolTimeout int
}

// MySQL config
type MySQL struct {
	MysqlHost       string
	MysqlPort       string
	MysqlUser       string
	MysqlPassword   string
	MysqlDBName     string
	MysqlUseSSL     bool
	MysqlCACertPath string
}

// GRPC gRPC service config
type GRPC struct {
	Host              string
	Port              string
	MaxConnectionIdle time.Duration
	Timeout           time.Duration
	MaxConnectionAge  time.Duration
}

// NATS Config
type Nats struct {
	Host string
	Port string
}

// HTTP Config
type Http struct {
	BaseUrl                 string
	Host                    string
	Port                    string
	OAuthTokenExpiresIn     int
	OAuthLongTokenExpiresIn int
	APP_REPORTS             string
}

type SMTP struct {
	SMTP_HOST     string
	SMTP_PORT     int32
	SMTP_FROM     string
	SMTP_PASSWORD string
	SMTP_LOGIN    string
}


// NewConfig get config from env
func NewConfig() *Config {
	// init config
	http := Http{}
	setting := Setting{}
	setting.LocalPath = "./locales/*/*"
	setting.Version = "1.0.0"
	logger := Logger{}
	logger.LogFile = "greenlync-api-gateway.log"
	redis := Redis{}
	gprc := GRPC{}
	mysql := MySQL{}
	nats := Nats{}
	smtp := SMTP{}

	c := &Config{
		HTTP:          http,
		GRPC:          gprc,
		Logger:        logger,
		Redis:         redis,
		MySQL:         mysql,
		Nats:          nats,
		Smtp:          smtp,
	}

	parseError := map[string]string{
		BASE_URL:                    "",
		LOGS_FILE:                   "",
		HTTP_HOST:                   "",
		HTTP_PORT:                   "",
		OAUTH_TOKEN_EXPIRES_IN:      "",
		OAUTH_LONG_TOKEN_EXPIRES_IN: "",
		APP_REPORTS:                 "",
		GRPC_HOST:                   "",
		GRPC_PORT:                   "",
		REDIS_URL:                   "",
		REDIS_PASSWORD:              "",
		MYSQL_HOST:                  "",
		MYSQL_PORT:                  "",
		MYSQL_USER:                  "",
		MYSQL_PASSWORD:              "",
		MYSQL_DB:                    "",
		NATS_HOST:                   "",
		NATS_PORT:                   "",
		SMTP_HOST:                   "",
		SMTP_PORT:                   "",
		SMTP_FROM:                   "",
		SMTP_PASSWORD:               "",
	}


	baseUrl := os.Getenv(BASE_URL)
	if baseUrl != "" {
		c.HTTP.BaseUrl = baseUrl
		parseError[BASE_URL] = baseUrl
	}

	logsFile := os.Getenv(LOGS_FILE)
	if logsFile != "" {
		c.Logger.LogFile = logsFile
		parseError[LOGS_FILE] = logsFile
	}

	httpHost := os.Getenv(HTTP_HOST)
	if httpHost != "" {
		c.HTTP.Host = httpHost
		parseError[HTTP_HOST] = httpHost
	}

	httpPort := os.Getenv(HTTP_PORT)
	if httpPort != "" {
		c.HTTP.Port = httpPort
		parseError[HTTP_PORT] = httpPort
	}

	oauthTokenExpiresIn, err := strconv.ParseInt(os.Getenv(OAUTH_TOKEN_EXPIRES_IN), 10, 64)
	if err == nil {
		c.HTTP.OAuthTokenExpiresIn = int(oauthTokenExpiresIn)
		parseError[OAUTH_TOKEN_EXPIRES_IN] = OAUTH_TOKEN_EXPIRES_IN
	}

	oAuthLongTokenExpiresIn, err := strconv.ParseInt(os.Getenv(OAUTH_LONG_TOKEN_EXPIRES_IN), 10, 64)
	if err == nil {
		c.HTTP.OAuthLongTokenExpiresIn = int(oAuthLongTokenExpiresIn)
		parseError[OAUTH_LONG_TOKEN_EXPIRES_IN] = OAUTH_LONG_TOKEN_EXPIRES_IN
	}

	reportData := os.Getenv(APP_REPORTS)
	if httpPort != "" {
		c.HTTP.APP_REPORTS = reportData
		parseError[APP_REPORTS] = reportData
	}

	redisURL := os.Getenv(REDIS_URL)
	if redisURL != "" {
		c.Redis.RedisAddr = redisURL
		parseError[REDIS_URL] = redisURL
	}

	redisPassword := os.Getenv(REDIS_PASSWORD)
	if redisPassword != "" {
		c.Redis.RedisPassword = redisPassword
		parseError[REDIS_PASSWORD] = redisPassword
	}

	gRPCHost := os.Getenv(GRPC_HOST)
	if gRPCHost != "" {
		c.GRPC.Host = gRPCHost
		parseError[GRPC_HOST] = gRPCHost
	}

	gRPCPort := os.Getenv(GRPC_PORT)
	if gRPCPort != "" {
		c.GRPC.Port = gRPCPort
		parseError[GRPC_PORT] = gRPCPort

	}

	mysqlHost := os.Getenv(MYSQL_HOST)
	if mysqlHost != "" {
		c.MySQL.MysqlHost = mysqlHost
		parseError[MYSQL_HOST] = mysqlHost
	}

	mysqlPort := os.Getenv(MYSQL_PORT)
	if mysqlPort != "" {
		c.MySQL.MysqlPort = mysqlPort
		parseError[MYSQL_PORT] = mysqlPort

	}

	mysqlUser := os.Getenv(MYSQL_USER)
	if mysqlUser != "" {
		c.MySQL.MysqlUser = mysqlUser
		parseError[MYSQL_USER] = mysqlUser
	}

	mysqlPassword := os.Getenv(MYSQL_PASSWORD)
	if mysqlPassword != "" {
		c.MySQL.MysqlPassword = mysqlPassword
		parseError[MYSQL_PASSWORD] = mysqlPassword
	}

	mysqlDBName := os.Getenv(MYSQL_DB)
	if mysqlDBName != "" {
		c.MySQL.MysqlDBName = mysqlDBName
		parseError[MYSQL_DB] = mysqlDBName
	}

	mysqlUseSSL, err := strconv.ParseBool(os.Getenv(MYSQL_USE_SSL))
	if err == nil {
		c.MySQL.MysqlUseSSL = mysqlUseSSL
		parseError[MYSQL_USE_SSL] = MYSQL_USE_SSL
	}

	mysqlCACertPath := os.Getenv(MYSQL_CACERT_PATH)
	if mysqlCACertPath != "" {
		c.MySQL.MysqlCACertPath = mysqlCACertPath
		parseError[MYSQL_CACERT_PATH] = mysqlCACertPath
	}

	natsHost := os.Getenv(NATS_HOST)
	if natsHost != "" {
		c.Nats.Host = natsHost
		parseError[NATS_HOST] = natsHost
	}

	natsPort := os.Getenv(NATS_PORT)
	if natsPort != "" {
		c.Nats.Port = natsPort
		parseError[NATS_PORT] = natsPort
	}

	smtpHost := os.Getenv(SMTP_HOST)
	if smtpHost != "" {
		c.Smtp.SMTP_HOST = smtpHost
		parseError[SMTP_HOST] = SMTP_HOST
	}

	smtpPort, err := strconv.ParseInt(os.Getenv(SMTP_PORT), 10, 32)
	if err == nil {
		c.Smtp.SMTP_PORT = int32(smtpPort)
		parseError[SMTP_PORT] = SMTP_PORT
	}

	smtpFrom := os.Getenv(SMTP_FROM)
	if smtpFrom != "" {
		c.Smtp.SMTP_FROM = smtpFrom
		parseError[SMTP_FROM] = smtpFrom
	}

	smtpPassword := os.Getenv(SMTP_PASSWORD)
	if smtpPassword != "" {
		c.Smtp.SMTP_PASSWORD = smtpPassword
		parseError[SMTP_PASSWORD] = SMTP_PASSWORD
	}

	smtpLogin := os.Getenv(SMTP_LOGIN)
	if smtpLogin != "" {
		c.Smtp.SMTP_LOGIN = smtpLogin
		parseError[SMTP_LOGIN] = SMTP_LOGIN
	}

	exitParse := false
	for k, v := range parseError {
		if v == "" {
			exitParse = true
			fmt.Printf("%s = %s\n", k, v)
		}
	}

	// one faild
	if exitParse {
		panic("Env vars not set see list")
	}
	return c
}
