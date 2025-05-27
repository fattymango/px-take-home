package server

import (
	"os"

	"github.com/fattymango/px-take-home/internal/middleware"
	"github.com/gofiber/fiber/v2"

	"github.com/gofiber/contrib/swagger"
)

func (s *Server) RegisterRoutes() {

	root := s.App.Group("/")

	// CORS Middleware
	root.Use(middleware.CORS())

	// Logging Middleware
	root.Use(middleware.Logger(s.logger))

	// Rate Limiter Middleware
	// root.Use(middleware.RateLimiter(s.logger))

	// Swagger UI
	s.RegisterSwagger(root)

	// Serve static web client files
	root.Static("/", "./web")

	api := root.Group("/api")
	v1 := api.Group("/v1")

	// Task
	s.RegisterTaskAPIs(v1)

	// SSE
	s.RegisterSSEHandlers(v1)

}

func (s *Server) RegisterSwagger(router fiber.Router) {
	swaggerPath := s.config.Swagger.FilePath
	// Check if the file exists
	if _, err := os.Stat(swaggerPath); os.IsNotExist(err) {
		s.logger.Error("Swagger file not found, swagger will not be available")
		return
	}

	router.Use(swagger.New(swagger.Config{
		FilePath: s.config.Swagger.FilePath,
		Path:     "/api/v1/docs",
		CacheAge: 1,
	}))
}

func (s *Server) RegisterTaskAPIs(router fiber.Router) {
	task := router.Group("/tasks")

	task.Post("/", s.CreateTask)
	task.Get("/", s.GetAllTasks)
	task.Get("/:taskID", s.GetTaskByID)
	task.Get("/:taskID/logs", s.GetTaskLogsByID)
	task.Get("/:taskID/logs/download", s.DownloadTaskLogs)
	task.Delete("/:taskID/cancel", s.CancelTask)
}

func (s *Server) RegisterSSEHandlers(router fiber.Router) error {
	router.Get("/events", s.SSE)

	return nil
}
