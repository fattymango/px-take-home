package middleware

import (
	"time"

	"github.com/fattymango/px-take-home/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func RateLimiter(logger *logger.Logger) fiber.Handler {
	return limiter.New(limiter.Config{
		LimiterMiddleware: limiter.SlidingWindow{},
		Max:               100,             // maximum requests allowed per duration
		Expiration:        1 * time.Minute, // duration for which requests are counted
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded, please try again later",
			})
		},
	})

}
