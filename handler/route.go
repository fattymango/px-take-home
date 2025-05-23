package server

import (
	// _ "github.com/fattymango/px-take-home/api/swagger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	// fiberSwagger "github.com/swaggo/fiber-swagger"
	"github.com/gofiber/contrib/swagger"
)

/*
GUIDE FOR ADDING NEW ROUTES:
1. If the request is in the context of a business, like creating a client, then add the route like this:
- POST /{businessID}/client/ , where we specify the businessID first, then the resource, then the action.
2. If the request is a simple create,update,delete or get, like GetAllGlient, then add the route like this:
- GET /{businessID}/client/
3. If the request is a simple get, like GetClientByID, then add the route like this:
- GET /{businessID}/client/{clientID}
4. If its not a regular CRUD operation, then add the route like this for example:
- POST /{businessID}/lead/convert/{leadID}, try to make the route as descriptive as possible.
*/
func (s *Server) RegisterRoutes() {

	root := s.App.Group("/")

	// CORS Middleware
	root.Use(cors.New())

	// Loging Middleware
	root.Use(s.middlewares.Logger)
	root.Use(s.middlewares.RateLimiter)
	// Swagger UI
	root.Use(swagger.New(swagger.Config{
		FilePath: s.config.Swagger.FilePath,
		Path:     "/api/v1/docs",
		CacheAge: 1,
	}))

	root.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	api := root.Group("/api")
	v1 := api.Group("/v1")

	// Health
	s.RegisterHealthAPIs(v1)

	// Task
	s.RegisterTaskAPIs(v1)

}

func (s *Server) RegisterHealthAPIs(router fiber.Router) {
	health := router.Group("/health")

	health.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
}

func (s *Server) RegisterTaskAPIs(router fiber.Router) {
	task := router.Group("/tasks")

	task.Post("/", s.CreateTask)
	task.Get("/", s.GetAllTasks)
	task.Get("/:taskID", s.GetTaskByID)
}
