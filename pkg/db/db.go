package db

import (
	"time"

	"github.com/fattymango/px-take-home/config" // Sqlite driver based on CGO
	"gorm.io/gorm"
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
