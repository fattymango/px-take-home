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

func GetOffsetLimitFromCtx(ctx *fiber.Ctx) (int, int) {
	offset := ctx.QueryInt("offset", 0)
	limit := ctx.QueryInt("limit", 10)

	if limit > 100 {
		limit = 100
	}

	return offset, limit
}
