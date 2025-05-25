package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fattymango/px-take-home/config"
	server "github.com/fattymango/px-take-home/handler"

	"github.com/fattymango/px-take-home/pkg/db"
	"github.com/fattymango/px-take-home/pkg/logger"
)

func Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// Create server
	log.Info("Creating server...")
	s, err := server.NewServer(cfg, log, db)
	if err != nil {
		log.Fatalf("failed to create server: %s", err)
	}
	log.Info("Server Created Successfully")

	// Start server
	log.Info("Starting Server")
	go func() {
		err = s.Start()
		if err != nil {
			log.Fatalf("failed to start server: %s", err)
		}
	}()

	// when painc receover
	if err := recover(); err != nil {
		log.Fatalf("some panic ...:", err)
	}

	// we need nice way to exit will use os package notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case v := <-quit:
		log.Infof("signal.Notify CTRL+C: %v", v)
	case done := <-ctx.Done():
		log.Infof("ctx.Done: %v", done)
	}

	log.Info("Stopping server...")
	err = s.Stop()
	if err != nil {
		log.Fatalf("failed to stop server: %s", err)
	}
	log.Info("Server stopped")

}
