package middleware

import (
	"time"

	"github.com/fattymango/px-take-home/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

func Logger(logger *logger.Logger) fiber.Handler {
	return func(ctx *fiber.Ctx) error {

		start := time.Now()
		err := ctx.Next()
		duration := time.Since(start)

		if err != nil {
			logger.Errorf("Error: %s", err.Error())
		}

		logger.Infof("%d | %s | %s | %v", ctx.Response().StatusCode(), ctx.Method(), ctx.Path(), duration)
		return err
	}
}
