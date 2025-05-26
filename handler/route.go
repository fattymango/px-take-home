package server

import (
	// _ "github.com/fattymango/px-take-home/api/swagger"

	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	// fiberSwagger "github.com/swaggo/fiber-swagger"
	"github.com/gofiber/contrib/swagger"
)

func (s *Server) RegisterRoutes() {

	root := s.App.Group("/")

	// CORS Middleware
	root.Use(cors.New())

	// Loging Middleware
	root.Use(s.middlewares.Logger)
	// root.Use(s.middlewares.RateLimiter)

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
	task.Delete("/:taskID/cancel", s.CancelTask)
}

func (s *Server) RegisterSSEHandlers(router fiber.Router) error {
	router.Get("/events", s.SSE)

	return nil
}
