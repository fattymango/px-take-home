package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fattymango/px-take-home/config"
	"github.com/stretchr/testify/assert"
)

func TestNewSQLiteDB_Success(t *testing.T) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "sqlite_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test config
	cfg := &config.Config{
		DB: config.DB{
			File:            filepath.Join(tmpDir, "test.db"),
			MaxIdleConns:    2,
			MaxOpenConns:    5,
			MaxConnLifetime: 10,
		},
	}

	// Initialize database
	db, err := NewSQLiteDB(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Verify the database connection works
	sqlDB, err := db.DB.DB()
	assert.NoError(t, err)
	assert.NoError(t, sqlDB.Ping())
}

func TestNewSQLiteDB_InvalidPath(t *testing.T) {
	// Create config with invalid path
	cfg := &config.Config{
		DB: config.DB{
			File:            "/nonexistent/directory/that/we/cant/create/test.db",
			MaxIdleConns:    2,
			MaxOpenConns:    5,
			MaxConnLifetime: 10,
		},
	}

	// Try to initialize database
	db, err := NewSQLiteDB(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to create directory")
}
