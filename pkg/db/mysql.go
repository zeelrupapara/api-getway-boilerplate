// By Emran A. Hamdan, Lead Architect
// Developer: Saif Hamdan

package db

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"
	"greenlync-api-gateway/config"

	model "greenlync-api-gateway/model/common/v1"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

// MysqlDB instant to pass to handlers
type MysqlDB struct {
	DB *gorm.DB
}

// NewMysqlDB will return a valid connection to Mysql DB Session
func NewMysqDB(cfg *config.Config) (*MysqlDB, error) {
	var dsn string

	if cfg.MySQL.MysqlUseSSL {
		// Load CA certificate
		rootCertPool := x509.NewCertPool()
		pem, err := os.ReadFile(cfg.MySQL.MysqlCACertPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read CA cert: %w", err)
		}

		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("failed to append CA cert to pool")
		}

		tlsConfig := &tls.Config{
			RootCAs: rootCertPool,
		}

		// Register the TLS configuration
		err = mysqlDriver.RegisterTLSConfig("custom", tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to register custom TLS config: %w", err)
		}

		// Construct DSN with TLS
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=custom",
			cfg.MySQL.MysqlUser,
			cfg.MySQL.MysqlPassword,
			cfg.MySQL.MysqlHost,
			cfg.MySQL.MysqlPort,
			cfg.MySQL.MysqlDBName,
		)
	} else {
		// Construct DSN without TLS
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.MySQL.MysqlUser,
			cfg.MySQL.MysqlPassword,
			cfg.MySQL.MysqlHost,
			cfg.MySQL.MysqlPort,
			cfg.MySQL.MysqlDBName,
		)
	}

	fmt.Println("DSN", dsn)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true, // Disable automatic transactions for read-only operations

		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "vfx_",
			SingularTable: true,
			NoLowerCase:   true,
		},
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)                 // Maximum idle connections
	sqlDB.SetMaxOpenConns(1000)               // Maximum open connections
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Maximum connection lifetime

	dbc := &MysqlDB{
		DB: db,
	}

	return dbc, nil
}

// Migrate when you change your model, called from main only
func (db *MysqlDB) Migrate() error {
	// Core models for cannabis compliance
	if err := db.DB.AutoMigrate(&model.User{}, &model.Role{}, &model.Permission{}); err != nil {
		return err
	}
	// Authorization and sessions
	if err := db.DB.AutoMigrate(&model.Token{}, &model.Session{}); err != nil {
		return err
	}
	// Configuration management
	if err := db.DB.AutoMigrate(&model.ConfigGroup{}, &model.Config{}); err != nil {
		return err
	}
	// Communication
	if err := db.DB.AutoMigrate(&model.Mail{}, &model.Inbox{}); err != nil {
		return err
	}
	// Cannabis compliance
	if err := db.DB.AutoMigrate(&model.ComplianceCategory{}); err != nil {
		return err
	}
	// Events and monitoring
	if err := db.DB.AutoMigrate(&model.Event{}, &model.OperationsLog{}); err != nil {
		return err
	}

	return nil
}


func (db *MysqlDB) ValidateDBData() error {
	// Cannabis boilerplate - simplified validation
	// Check if we have at least one user
	var userCount int64
	err := db.DB.Model(&model.User{}).Count(&userCount).Error
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}
	
	if userCount == 0 {
		return fmt.Errorf("no users found - please run database seeding")
	}
	
	// Check if we have at least one admin user
	var adminCount int64
	err = db.DB.Model(&model.User{}).Where("role = ?", "admin").Count(&adminCount).Error
	if err != nil {
		return fmt.Errorf("failed to count admin users: %w", err)
	}
	
	if adminCount == 0 {
		return fmt.Errorf("no admin users found - please run database seeding")
	}

	return nil
}

// AlterTableIds - Cannabis boilerplate stub (no longer needed)
func (db *MysqlDB) AlterTableIds() error {
	// This function is not needed for cannabis compliance system
	// Keeping as stub for backward compatibility
	return nil
}
