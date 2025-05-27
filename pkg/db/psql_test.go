package db

import (
	"os"
	"testing"
	"time"

	"github.com/fattymango/px-take-home/config"
	"github.com/stretchr/testify/assert"
)

func TestNewPsqlDB_Success(t *testing.T) {
	// Use environment variables or default test values
	cfg := &config.Config{
		DB: config.DB{
			Host:            getEnvOrDefault("POSTGRES_HOST", "localhost"),
			Port:            getEnvOrDefault("POSTGRES_PORT", "5432"),
			User:            getEnvOrDefault("POSTGRES_USER", "postgres"),
			Password:        getEnvOrDefault("POSTGRES_PASSWORD", "postgres"),
			Name:            getEnvOrDefault("POSTGRES_DB", "postgres"),
			SSLMode:         "disable",
			MaxIdleConns:    2,
			MaxOpenConns:    5,
			MaxConnLifetime: 10,
		},
	}

	// Initialize database
	db, err := NewPsqlDB(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Verify the database connection works
	sqlDB, err := db.DB.DB()
	assert.NoError(t, err)
	assert.NoError(t, sqlDB.Ping())

	// Close the connection
	assert.NoError(t, sqlDB.Close())
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestNewPsqlDB_ConnectionFailure(t *testing.T) {
	// Create config with invalid connection details
	cfg := &config.Config{
		DB: config.DB{
			Host:            "nonexistent-host",
			Port:            "5432",
			User:            "testuser",
			Password:        "testpass",
			Name:            "testdb",
			SSLMode:         "disable",
			MaxIdleConns:    2,
			MaxOpenConns:    5,
			MaxConnLifetime: 10,
		},
	}

	// Try to initialize database - should fail after retries
	startTime := time.Now()
	db, err := NewPsqlDB(cfg)
	duration := time.Since(startTime)

	// Verify error handling
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to connect to PostgreSQL after retries")

	// Verify that retries occurred (should take around 3 * 2 seconds due to retry logic)
	assert.GreaterOrEqual(t, duration, 6*time.Second, "Should have attempted retries")
}

func TestNewPsqlDB_InvalidConfig(t *testing.T) {
	// Test cases with invalid configurations
	testCases := []struct {
		name        string
		config      config.DB
		errorString string
	}{
		{
			name: "Empty Host",
			config: config.DB{
				Host:            "",
				Port:            "5432",
				User:            "testuser",
				Password:        "testpass",
				Name:            "testdb",
				SSLMode:         "disable",
				MaxIdleConns:    2,
				MaxOpenConns:    5,
				MaxConnLifetime: 10,
			},
			errorString: "failed to connect to PostgreSQL after retries",
		},
		{
			name: "Invalid Port",
			config: config.DB{
				Host:            "localhost",
				Port:            "invalid",
				User:            "testuser",
				Password:        "testpass",
				Name:            "testdb",
				SSLMode:         "disable",
				MaxIdleConns:    2,
				MaxOpenConns:    5,
				MaxConnLifetime: 10,
			},
			errorString: "failed to connect to PostgreSQL after retries",
		},
		{
			name: "Empty Database Name",
			config: config.DB{
				Host:            "localhost",
				Port:            "5432",
				User:            "testuser",
				Password:        "testpass",
				Name:            "",
				SSLMode:         "disable",
				MaxIdleConns:    2,
				MaxOpenConns:    5,
				MaxConnLifetime: 10,
			},
			errorString: "failed to connect to PostgreSQL after retries",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				DB: tc.config,
			}

			db, err := NewPsqlDB(cfg)
			assert.Error(t, err)
			assert.Nil(t, db)
			assert.Contains(t, err.Error(), tc.errorString)
		})
	}
}
