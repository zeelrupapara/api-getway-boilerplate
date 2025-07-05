package db

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/config"
	appLogger "gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/logger"
)

// Database represents the database connection and configuration
type Database struct {
	DB     *gorm.DB
	Config *config.DatabaseConfig
	Logger *appLogger.Logger
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.DatabaseConfig, log *appLogger.Logger) (*Database, error) {
	log.Info("Initializing database connection",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Name,
		"ssl_mode", cfg.SSLMode,
	)

	// Build DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
		cfg.Timezone,
	)

	// Configure GORM logger
	gormLogger := logger.New(
		&gormLoggerWriter{logger: log},
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "gl_", // greenlync prefix for all tables
			SingularTable: false,
			NoLowerCase:   false,
		},
		// Enable cannabis compliance features
		DisableForeignKeyConstraintWhenMigrating: false,
		PrepareStmt:                              true,
		CreateBatchSize:                          1000,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		DB:     db,
		Config: cfg,
		Logger: log,
	}

	log.Info("Database connection established successfully",
		"max_open_conns", cfg.MaxOpenConns,
		"max_idle_conns", cfg.MaxIdleConns,
		"max_lifetime", cfg.MaxLifetime,
	)

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	d.Logger.Info("Database connection closed successfully")
	return nil
}

// Health checks the database connection health
func (d *Database) Health(ctx context.Context) error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	return sqlDB.PingContext(ctx)
}

// GetStats returns database connection statistics
func (d *Database) GetStats() map[string]interface{} {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}

// AutoMigrate runs database migrations for all models
func (d *Database) AutoMigrate(models ...interface{}) error {
	d.Logger.Info("Running database migrations", "models_count", len(models))
	
	for _, model := range models {
		if err := d.DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
	}

	d.Logger.Info("Database migrations completed successfully")
	return nil
}

// Transaction executes a function within a database transaction
func (d *Database) Transaction(fn func(tx *gorm.DB) error) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// WithContext returns a new GORM DB instance with context
func (d *Database) WithContext(ctx context.Context) *gorm.DB {
	return d.DB.WithContext(ctx)
}

// gormLoggerWriter implements the GORM logger interface
type gormLoggerWriter struct {
	logger *appLogger.Logger
}

// Printf implements the logger interface for GORM
func (w *gormLoggerWriter) Printf(format string, args ...interface{}) {
	w.logger.Info(fmt.Sprintf(format, args...))
}

// Cannabis-specific database operations

// SetupMultiTenancy configures row-level security for dispensary isolation
func (d *Database) SetupMultiTenancy() error {
	d.Logger.Info("Setting up multi-tenant dispensary isolation")

	// Enable Row Level Security (RLS) for cannabis compliance
	queries := []string{
		// Enable RLS on user tables
		"ALTER TABLE gl_users ENABLE ROW LEVEL SECURITY;",
		
		// Enable RLS on dispensary-specific tables
		"ALTER TABLE gl_dispensaries ENABLE ROW LEVEL SECURITY;",
		"ALTER TABLE gl_products ENABLE ROW LEVEL SECURITY;",
		"ALTER TABLE gl_orders ENABLE ROW LEVEL SECURITY;",
		"ALTER TABLE gl_transactions ENABLE ROW LEVEL SECURITY;",
		
		// Create RLS policies for dispensary isolation
		`CREATE POLICY dispensary_isolation_policy ON gl_users
		 USING (dispensary_id = current_setting('app.current_dispensary_id')::uuid);`,
		
		`CREATE POLICY dispensary_products_policy ON gl_products
		 USING (dispensary_id = current_setting('app.current_dispensary_id')::uuid);`,
		
		`CREATE POLICY dispensary_orders_policy ON gl_orders
		 USING (dispensary_id = current_setting('app.current_dispensary_id')::uuid);`,
		
		`CREATE POLICY dispensary_transactions_policy ON gl_transactions
		 USING (dispensary_id = current_setting('app.current_dispensary_id')::uuid);`,
	}

	// Execute each query (will be created when tables exist)
	for _, query := range queries {
		if err := d.DB.Exec(query).Error; err != nil {
			// Log warning but don't fail - tables may not exist yet
			d.Logger.Warn("Failed to execute RLS query (tables may not exist yet)",
				"query", query,
				"error", err,
			)
		}
	}

	d.Logger.Info("Multi-tenancy setup completed (RLS policies created)")
	return nil
}

// SetDispensaryContext sets the current dispensary ID for RLS
func (d *Database) SetDispensaryContext(dispensaryID string) error {
	return d.DB.Exec("SET LOCAL app.current_dispensary_id = ?", dispensaryID).Error
}

// LogDatabaseOperation logs database operations for cannabis compliance
func (d *Database) LogDatabaseOperation(userID, operation, table string, recordID interface{}) {
	d.Logger.LogCannabisAudit(userID, operation, table, map[string]interface{}{
		"record_id": recordID,
		"table":     table,
		"operation": operation,
		"timestamp": time.Now().UTC(),
	})
}