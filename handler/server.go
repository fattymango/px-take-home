package server

import (
	"fmt"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/internal/middleware"
	"github.com/fattymango/px-take-home/internal/sse"
	"github.com/fattymango/px-take-home/internal/task"
	"github.com/fattymango/px-take-home/pkg/db"
	"github.com/fattymango/px-take-home/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Server struct {
	config    *config.Config
	logger    *logger.Logger
	App       *fiber.App
	db        *db.DB
	validator *validator.Validate

	TaskManager *task.TaskManager

	sseManager *sse.SseManager
}

func NewServer(cfg *config.Config, logger *logger.Logger, db *db.DB) (*Server, error) {
	taskManager := task.NewTaskManager(cfg, logger, task.NewTaskDBStore(cfg, logger, db))

	return &Server{
		config:      cfg,
		logger:      logger,
		App:         fiber.New(),
		db:          db,
		validator:   validator.New(),
		TaskManager: taskManager,
		sseManager:  sse.NewSseManager(cfg, logger, taskManager.TaskUpdatesStream(), taskManager.LogStream()),
	}, nil
}

func (s *Server) Start() error {
	s.logger.Info("Starting server...")
	s.RegisterRoutes()
	s.TaskManager.Start()
	s.sseManager.Start()
	err := s.App.Listen(":" + s.config.Server.Port)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func (s *Server) Stop() error {
	s.logger.Info("Stopping server...")
	s.TaskManager.Stop()
	s.sseManager.Stop()
	err := s.App.Shutdown()

	return err
}

type Middlewares struct {
	RateLimiter fiber.Handler
	Logger      fiber.Handler
}

func newMiddlewares(logger *logger.Logger) *Middlewares {
	return &Middlewares{
		RateLimiter: middleware.RateLimiter(logger),
		Logger:      middleware.Logger(logger),
	}
}
