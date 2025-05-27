package ctxstore

import (
	"fmt"

	"github.com/fattymango/px-take-home/model"
	"github.com/gofiber/fiber/v2"
)

func GetTaskIDFromCtx(ctx *fiber.Ctx) (string, error) {
	taskID := ctx.Params("taskID")
	if taskID == "" {
		return "", fmt.Errorf("taskID is required")
	}

	return taskID, nil
}

func GetOffsetLimitQueryFromCtx(ctx *fiber.Ctx) (int, int) {
	offset := ctx.QueryInt("offset", 0)
	limit := ctx.QueryInt("limit", 10)

	if limit > 100 {
		limit = 100
	}

	return offset, limit
}

func GetStatusQueryFromCtx(ctx *fiber.Ctx) (model.TaskStatus, error) {
	status := ctx.QueryInt("status", 0)

	if status == 0 {
		return 0, nil
	}

	if _, ok := model.TaskStatus_name[model.TaskStatus(status)]; !ok {
		return 0, fmt.Errorf("invalid status")
	}

	return model.TaskStatus(status), nil
}
