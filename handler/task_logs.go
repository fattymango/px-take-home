package server

import (
	"github.com/fattymango/px-take-home/dto"
	"github.com/fattymango/px-take-home/pkg/ctxstore"
	"github.com/gofiber/fiber/v2"
)

// @Tags Task Logs
// @Summary Get task logs by ID
// @Router /api/v1/tasks/{taskID}/logs [get]
// @Security BearerAuth
// @Description Get task logs by ID
// @Accept json
// @Produce json
// @Param taskID path int true "Task ID"
// @Param filter body dto.TaskLogFilter true "Filter"
//
//	@Success	200	{object} dto.ViewTaskLogs "Success"
//	@Failure	400	{object} dto.BaseResponse	"Bad Request"
//	@Failure	401	{object} dto.BaseResponse	"Unauthorized"
//	@Failure	404	{object} dto.BaseResponse	"Not Found"
//	@Failure	500	{object} dto.BaseResponse	"Internal Server Error"
//
// @Security BearerAuth
// @ID GetTaskLogsByID
func (s *Server) GetTaskLogsByID(c *fiber.Ctx) error {
	taskID, err := ctxstore.GetTaskIDFromCtx(c)
	if err != nil {
		return dto.NewBadRequestResponse(c, err.Error())
	}

	filter := &dto.TaskLogFilter{}
	if err := c.QueryParser(filter); err != nil {
		return dto.NewBadRequestResponse(c, err.Error())
	}

	if filter.From != 0 && filter.To != 0 && filter.From >= filter.To {
		return dto.NewBadRequestResponse(c, "from must be less than to")
	}

	logs, totalLines, err := s.TaskManager.GetTaskLogs(taskID, filter.From, filter.To)
	if err != nil {
		return dto.NewInternalServerErrorResponse(c, err.Error())
	}

	return dto.NewSuccessResponse(c, dto.ToViewTaskLogs(logs, totalLines))
}
