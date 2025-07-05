package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the GreenLync API Gateway
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	NATS     NATSConfig     `mapstructure:"nats"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Cannabis CannabisConfig `mapstructure:"cannabis"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host           string        `mapstructure:"host" default:"0.0.0.0"`
	Port           int           `mapstructure:"port" default:"8080"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout" default:"30s"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout" default:"30s"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout" default:"120s"`
	AllowedOrigins string        `mapstructure:"allowed_origins" default:"*"`
	Debug          bool          `mapstructure:"debug" default:"false"`
}

// DatabaseConfig holds PostgreSQL database configuration
type DatabaseConfig struct {
	Host         string `mapstructure:"host" default:"localhost"`
	Port         int    `mapstructure:"port" default:"5432"`
	User         string `mapstructure:"user" default:"postgres"`
	Password     string `mapstructure:"password"`
	Name         string `mapstructure:"name" default:"greenlync"`
	SSLMode      string `mapstructure:"ssl_mode" default:"disable"`
	Timezone     string `mapstructure:"timezone" default:"UTC"`
	MaxOpenConns int    `mapstructure:"max_open_conns" default:"25"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" default:"10"`
	MaxLifetime  int    `mapstructure:"max_lifetime" default:"3600"`
}

// RedisConfig holds Redis configuration for session management
type RedisConfig struct {
	Host        string        `mapstructure:"host" default:"localhost"`
	Port        int           `mapstructure:"port" default:"6379"`
	Password    string        `mapstructure:"password"`
	Database    int           `mapstructure:"database" default:"0"`
	MaxRetries  int           `mapstructure:"max_retries" default:"3"`
	PoolSize    int           `mapstructure:"pool_size" default:"10"`
	MinIdleConn int           `mapstructure:"min_idle_conn" default:"5"`
	DialTimeout time.Duration `mapstructure:"dial_timeout" default:"5s"`
	ReadTimeout time.Duration `mapstructure:"read_timeout" default:"3s"`
}

// NATSConfig holds NATS messaging configuration
type NATSConfig struct {
	URL             string        `mapstructure:"url" default:"nats://localhost:4222"`
	ClusterID       string        `mapstructure:"cluster_id" default:"greenlync-cluster"`
	ClientID        string        `mapstructure:"client_id" default:"greenlync-gateway"`
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout" default:"10s"`
	ReconnectWait   time.Duration `mapstructure:"reconnect_wait" default:"2s"`
	MaxReconnects   int           `mapstructure:"max_reconnects" default:"10"`
	PingInterval    time.Duration `mapstructure:"ping_interval" default:"2m"`
	MaxPingsOut     int           `mapstructure:"max_pings_out" default:"2"`
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret           string        `mapstructure:"secret"`
	AccessExpiry     time.Duration `mapstructure:"access_expiry" default:"15m"`
	RefreshExpiry    time.Duration `mapstructure:"refresh_expiry" default:"168h"`
	Issuer           string        `mapstructure:"issuer" default:"greenlync-gateway"`
	Audience         string        `mapstructure:"audience" default:"greenlync-platform"`
	RequireHTTPS     bool          `mapstructure:"require_https" default:"false"`
	MaxSessions      int           `mapstructure:"max_sessions" default:"3"`
}

// CannabisConfig holds cannabis industry specific configuration
type CannabisConfig struct {
	AgeVerificationRequired bool     `mapstructure:"age_verification_required" default:"true"`
	MinimumAge             int      `mapstructure:"minimum_age" default:"21"`
	LegalStates            []string `mapstructure:"legal_states"`
	ComplianceMode         string   `mapstructure:"compliance_mode" default:"strict"`
	AuditLogging           bool     `mapstructure:"audit_logging" default:"true"`
	StateCheckEnabled      bool     `mapstructure:"state_check_enabled" default:"true"`
	PurchaseLimitTracking  bool     `mapstructure:"purchase_limit_tracking" default:"true"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level       string `mapstructure:"level" default:"info"`
	Format      string `mapstructure:"format" default:"json"`
	Output      string `mapstructure:"output" default:"stdout"`
	Structured  bool   `mapstructure:"structured" default:"true"`
	FileOutput  string `mapstructure:"file_output"`
	MaxFileSize int    `mapstructure:"max_file_size" default:"100"`
	MaxBackups  int    `mapstructure:"max_backups" default:"3"`
	MaxAge      int    `mapstructure:"max_age" default:"28"`
}

// NewConfig loads configuration from environment variables and config files
func NewConfig() (*Config, error) {
	// Set configuration paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("/etc/greenlync/")

	// Environment variable configuration
	viper.SetEnvPrefix("GREENLYNC")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is acceptable, continue with env vars and defaults
	}

	// Unmarshal configuration
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("server.allowed_origins", "*")
	viper.SetDefault("server.debug", false)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.name", "greenlync")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.timezone", "UTC")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_lifetime", 3600)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.database", 0)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conn", 5)
	viper.SetDefault("redis.dial_timeout", "5s")
	viper.SetDefault("redis.read_timeout", "3s")

	// NATS defaults
	viper.SetDefault("nats.url", "nats://localhost:4222")
	viper.SetDefault("nats.cluster_id", "greenlync-cluster")
	viper.SetDefault("nats.client_id", "greenlync-gateway")
	viper.SetDefault("nats.connect_timeout", "10s")
	viper.SetDefault("nats.reconnect_wait", "2s")
	viper.SetDefault("nats.max_reconnects", 10)
	viper.SetDefault("nats.ping_interval", "2m")
	viper.SetDefault("nats.max_pings_out", 2)

	// JWT defaults
	viper.SetDefault("jwt.access_expiry", "15m")
	viper.SetDefault("jwt.refresh_expiry", "168h")
	viper.SetDefault("jwt.issuer", "greenlync-gateway")
	viper.SetDefault("jwt.audience", "greenlync-platform")
	viper.SetDefault("jwt.require_https", false)
	viper.SetDefault("jwt.max_sessions", 3)

	// Cannabis defaults
	viper.SetDefault("cannabis.age_verification_required", true)
	viper.SetDefault("cannabis.minimum_age", 21)
	viper.SetDefault("cannabis.legal_states", []string{
		"CA", "CO", "WA", "OR", "NV", "AZ", "NY", "IL", "NJ", "VA", 
		"CT", "MT", "VT", "AK", "MA", "ME", "MI", "MD", "MO", "OH",
	})
	viper.SetDefault("cannabis.compliance_mode", "strict")
	viper.SetDefault("cannabis.audit_logging", true)
	viper.SetDefault("cannabis.state_check_enabled", true)
	viper.SetDefault("cannabis.purchase_limit_tracking", true)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
	viper.SetDefault("logging.structured", true)
	viper.SetDefault("logging.max_file_size", 100)
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 28)
}

// validateConfig validates the loaded configuration
func validateConfig(config *Config) error {
	// Validate JWT secret
	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if len(config.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	// Validate database password
	if config.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}

	// Validate server port
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// Validate cannabis configuration
	if config.Cannabis.MinimumAge < 18 {
		return fmt.Errorf("minimum age must be at least 18")
	}

	if len(config.Cannabis.LegalStates) == 0 {
		return fmt.Errorf("at least one legal state must be configured")
	}

	// Validate compliance mode
	validModes := []string{"strict", "normal", "lenient"}
	isValidMode := false
	for _, mode := range validModes {
		if config.Cannabis.ComplianceMode == mode {
			isValidMode = true
			break
		}
	}
	if !isValidMode {
		return fmt.Errorf("invalid compliance mode: %s (must be strict, normal, or lenient)", config.Cannabis.ComplianceMode)
	}

	return nil
}

// GetDSN returns the PostgreSQL connection string
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		d.Host, d.User, d.Password, d.Name, d.Port, d.SSLMode, d.Timezone)
}

// GetRedisAddr returns the Redis connection address
func (r *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// IsStateCompliant checks if a state allows cannabis
func (c *CannabisConfig) IsStateCompliant(state string) bool {
	state = strings.ToUpper(strings.TrimSpace(state))
	for _, legalState := range c.LegalStates {
		if strings.ToUpper(legalState) == state {
			return true
		}
	}
	return false
}

// GetServerAddress returns the complete server address
func (s *ServerConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}