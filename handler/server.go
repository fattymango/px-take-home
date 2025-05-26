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
	config      *config.Config
	logger      *logger.Logger
	App         *fiber.App
	db          *db.DB
	validator   *validator.Validate
	middlewares *Middlewares

	*Services
	sseManager *sse.SseManager
}

func NewServer(cfg *config.Config, logger *logger.Logger, db *db.DB) (*Server, error) {
	v := validator.New()

	services := newServices(cfg, logger, db)

	middlewares := &Middlewares{
		RateLimiter: middleware.RateLimiter(logger),
		Logger:      middleware.Logger(logger),
	}
	return &Server{
		config:      cfg,
		logger:      logger,
		App:         fiber.New(),
		db:          db,
		validator:   v,
		middlewares: middlewares,
		Services:    services,
		sseManager:  sse.NewSseManager(cfg, logger, services.TaskManager.TaskUpdatesStream(), services.TaskManager.LogStream()),
	}, nil
}

func (s *Server) Start() error {
	s.logger.Info("Starting server...")
	s.registerValidator()
	s.RegisterRoutes()
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

type Services struct {
	TaskManager *task.TaskManager
}

func newServices(cfg *config.Config, logger *logger.Logger, db *db.DB) *Services {
	taskManager := task.NewTaskManager(cfg, logger, task.NewTaskDB(cfg, logger, db))
	taskManager.Start()
	return &Services{
		TaskManager: taskManager,
	}
}

func (s *Server) WithTransaction() (*Services, func() error, func() error, error) {
	db, commit, rollback, err := s.db.NewTransaction()
	if err != nil {
		return nil, nil, nil, err
	}
	return newServices(s.config, s.logger, db), commit, rollback, nil
}
