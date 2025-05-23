package server

import (
	"fmt"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/internal/middleware"
	"github.com/fattymango/px-take-home/internal/task"
	"github.com/fattymango/px-take-home/pkg/cache"
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
	cache       *cache.Cache
	validator   *validator.Validate
	middlewares *Middlewares

	*Services
}

func NewServer(cfg *config.Config, logger *logger.Logger, db *db.DB, cache *cache.Cache) (*Server, error) {
	v := validator.New()

	services := newServices(cfg, logger, db, cache)

	middlewares := &Middlewares{
		RateLimiter: middleware.RateLimiter(logger),
		Logger:      middleware.Logger(logger),
	}
	return &Server{
		config:      cfg,
		logger:      logger,
		App:         fiber.New(),
		db:          db,
		cache:       cache,
		validator:   v,
		middlewares: middlewares,
		Services:    services,
	}, nil
}

func (s *Server) Start() error {
	s.logger.Info("Starting server...")
	s.registerValidator()
	s.RegisterRoutes()

	err := s.App.Listen(":" + s.config.Server.Port)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

type Middlewares struct {
	RateLimiter fiber.Handler
	Logger      fiber.Handler
}

type Services struct {
	TaskManager *task.TaskManager
}

func newServices(cfg *config.Config, logger *logger.Logger, db *db.DB, cache *cache.Cache) *Services {
	taskManager := task.NewTaskManager(cfg, logger, db)
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
	return newServices(s.config, s.logger, db, s.cache), commit, rollback, nil
}
