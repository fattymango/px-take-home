package ctxstore

import (
	"fmt"

	"github.com/fattymango/px-take-home/dto"
	"github.com/fattymango/px-take-home/model"
	"github.com/gofiber/fiber/v2"
)

func GetTaskIDFromCtx(ctx *fiber.Ctx) (uint64, error) {
	taskID, err := ctx.ParamsInt("taskID")
	if err != nil {
		return 0, fmt.Errorf("taskID is required")
	}

	return uint64(taskID), nil
}

func GetTaskLogFilterFromCtx(ctx *fiber.Ctx) (*dto.TaskLogFilter, error) {
	filter := &dto.TaskLogFilter{}
	if err := ctx.QueryParser(filter); err != nil {
		return nil, fmt.Errorf("failed to parse task log filter: %w", err)
	}

	return filter, nil
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
