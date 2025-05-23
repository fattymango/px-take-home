package ctxstore

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func GetTaskIDFromCtx(ctx *fiber.Ctx) (uint64, error) {
	taskID, err := ctx.ParamsInt("taskID")
	if err != nil {
		return 0, fmt.Errorf("taskID is required")
	}

	return uint64(taskID), nil
}
