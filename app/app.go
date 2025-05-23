package app

import (
	"fmt"

	"github.com/fattymango/px-take-home/config"
	server "github.com/fattymango/px-take-home/handler"

	"github.com/fattymango/px-take-home/pkg/cache"
	"github.com/fattymango/px-take-home/pkg/db"
	"github.com/fattymango/px-take-home/pkg/logger"
)

func Start() {

	cfg, err := config.NewConfig()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %s", err))
	}

	log, err := logger.NewLogger(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to create logger: %s", err))
	}

	// DB connection
	log.Info("Creating db connection...")
	db, err := db.NewSQLiteDB(cfg)
	if err != nil {
		log.Fatalf("failed to create db connection: %s", err)
	}
	log.Info("DB connection successful")

	// Migrate
	log.Info("Migrating...")
	err = Migrate(cfg, db)
	if err != nil {
		log.Fatalf("failed to migrate: %s", err)
	}
	log.Info("Migration successful")

	// Cache
	log.Info("Creating cache...")
	cache, err := cache.NewCache(cfg)
	if err != nil {
		log.Fatalf("failed to create cache: %s", err)
	}
	log.Info("Cache successful")

	// Create server
	log.Info("Creating server...")
	s, err := server.NewServer(cfg, log, db, cache)
	if err != nil {
		log.Fatalf("failed to create server: %s", err)
	}
	log.Info("Server Created Successfully")

	// Start server
	log.Info("Starting Server")
	err = s.Start()
	if err != nil {
		log.Fatalf("failed to start server: %s", err)
	}

}
