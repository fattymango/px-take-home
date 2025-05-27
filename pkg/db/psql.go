package db

import (
	"context"
	"fmt"
	"time"

	psql "gorm.io/driver/postgres"

	"github.com/fattymango/px-take-home/config"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// NewPsqlDB initializes a new PostgreSQL database connection with retries & timeout
func NewPsqlDB(cfg *config.Config) (*DB, error) {
	// Set a GORM logger with better logging levels
	gormLogger := logger.Default.LogMode(logger.Warn)

	var db *gorm.DB
	var err error

	// Retry connection with exponential backoff
	err = retryWithBackoff(3, 2*time.Second, func() error {
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode)
		db, err = gorm.Open(psql.Open(dsn), &gorm.Config{
			PrepareStmt:                              true,
			DisableForeignKeyConstraintWhenMigrating: true,
			TranslateError:                           true,
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
		return nil, fmt.Errorf("failed to connect to PostgreSQL after retries: %w", err)
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
