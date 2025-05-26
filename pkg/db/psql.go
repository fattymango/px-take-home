package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fattymango/px-take-home/config"
	"gorm.io/driver/sqlite" // Sqlite driver based on CGO
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// PsqlDB instance to pass to handlers
type DB struct {
	*gorm.DB
	config *config.Config
}

// retryWithBackoff retries a function with exponential backoff
func retryWithBackoff(attempts int, sleep time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(sleep)
		sleep *= 2 // Exponential backoff
	}
	return err
}

// NewSQLiteDB initializes a new PostgreSQL database connection with retries & timeout
func NewSQLiteDB(cfg *config.Config) (*DB, error) {
	// Set a GORM logger with better logging levels
	gormLogger := logger.Default.LogMode(logger.Silent)

	var db *gorm.DB
	var err error

	// Before opening the DB:
	dbPath := cfg.DB.File
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Retry connection with exponential backoff
	err = retryWithBackoff(3, 2*time.Second, func() error {
		db, err = gorm.Open(sqlite.Open(cfg.DB.File), &gorm.Config{
			PrepareStmt:                              true,
			DisableForeignKeyConstraintWhenMigrating: true,
			Logger:                                   gormLogger,
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   "px_",
				SingularTable: true,
				NoLowerCase:   false,
			},
		})
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite after retries: %w", err)
	}

	// Get the underlying SQL database connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL DB instance: %w", err)
	}

	// Optimize connection pool settings
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)                                    // Increased idle connections
	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)                                    // Adjust based on workload
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DB.MaxConnLifetime) * time.Minute) // Adjusted to prevent stale connections

	// Ensure connection is valid with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &DB{db, cfg}, nil
}

// NewTransaction creates a new db instance with a transaction
func (p *DB) NewTransaction() (*DB, func() error, func() error, error) {
	tx := p.Begin()
	if tx.Error != nil {
		return nil, nil, nil, tx.Error
	}
	commit := func() error {
		return tx.Commit().Error
	}
	rollback := func() error {
		return tx.Rollback().Error
	}
	return &DB{tx, p.config}, commit, rollback, nil
}
