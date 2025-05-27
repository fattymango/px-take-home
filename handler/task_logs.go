package server

import (
	"fmt"

	"github.com/fattymango/px-take-home/dto"
	logreader "github.com/fattymango/px-take-home/internal/log_reader"
	"github.com/fattymango/px-take-home/model"
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
//
// @Param taskID path int true "Task ID"
// @Param filter query dto.TaskLogFilter true "Filter"
//
// @Success	200	{object} dto.ViewTaskLogs "Success"
// @Failure	400	{object} dto.BaseResponse	"Bad Request"
// @Failure	401	{object} dto.BaseResponse	"Unauthorized"
// @Failure	404	{object} dto.BaseResponse	"Not Found"
// @Failure	500	{object} dto.BaseResponse	"Internal Server Error"
//
// @Security BearerAuth
// @ID GetTaskLogsByID
func (s *Server) GetTaskLogsByID(c *fiber.Ctx) error {
	taskID, err := ctxstore.GetTaskIDFromCtx(c)
	if err != nil {
		return dto.NewBadRequestResponse(c, fmt.Sprintf("failed to get task ID: %s", err))
	}

	_, err = s.TaskManager.GetTask(taskID)
	if err != nil {
		return dto.NewNotFoundResponse(c, fmt.Sprintf("task #%d not found", taskID))
	}

	filter, err := ctxstore.GetTaskLogFilterFromCtx(c)
	if err != nil {
		return dto.NewBadRequestResponse(c, fmt.Sprintf("failed to get task log filter: %s", err))
	}

	if (filter.From > 0 && filter.To == 0) ||
		(filter.From == 0 && filter.To > 0) {
		return dto.NewBadRequestResponse(c, "to and from must be provided together")
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

// @Tags Task Logs
// @Summary Download task logs
// @Router /api/v1/tasks/{taskID}/logs/download [get]
// @Security BearerAuth
// @Description Download all logs for a task
// @Accept json
// @Produce text/plain
// @Param taskID path int true "Task ID"
//
//	@Success	200	{file} file "Log file"
//	@Failure	400	{object} dto.BaseResponse	"Bad Request"
//	@Failure	401	{object} dto.BaseResponse	"Unauthorized"
//	@Failure	404	{object} dto.BaseResponse	"Not Found"
//	@Failure	500	{object} dto.BaseResponse	"Internal Server Error"
//
// @Security BearerAuth
// @ID DownloadTaskLogs
func (s *Server) DownloadTaskLogs(c *fiber.Ctx) error {
	taskID, err := ctxstore.GetTaskIDFromCtx(c)
	if err != nil {
		return dto.NewBadRequestResponse(c, err.Error())
	}

	task, err := s.TaskManager.GetTask(taskID)
	if err != nil {
		return dto.NewNotFoundResponse(c, fmt.Sprintf("task #%d not found", taskID))
	}

	// Only allow downloading logs for completed, failed, or canceled tasks
	if task.Status == model.TaskStatus_Running || task.Status == model.TaskStatus_Queued {
		return dto.NewBadRequestResponse(c, "cannot download logs for running or queued tasks")
	}

	logFilePath := logreader.FormatFileName(s.config.TaskLogger.DirPath, taskID)

	if !logreader.CheckFileExists(logFilePath) {
		return dto.NewNotFoundResponse(c, "log file not found")
	}

	c.Set("Content-Type", "text/plain")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="task-%d.log"`, taskID))

	// Send the file
	return c.SendFile(logFilePath)
}
