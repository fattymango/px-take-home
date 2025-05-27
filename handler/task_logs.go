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

	// Get task to check its status
	task, err := s.TaskManager.GetTask(taskID)
	if err != nil {
		return dto.NewInternalServerErrorResponse(c, err.Error())
	}

	// Only allow downloading logs for completed, failed, or canceled tasks
	if task.Status == model.TaskStatus_Running || task.Status == model.TaskStatus_Queued {
		return dto.NewBadRequestResponse(c, "cannot download logs for running or queued tasks")
	}

	// Get the log file path
	logFilePath := logreader.FormatFileName(s.config.TaskLogger.DirPath, taskID)

	// Check if the file exists
	if !logreader.CheckFileExists(logFilePath) {
		return dto.NewNotFoundResponse(c, "log file not found")
	}

	// Set response headers
	c.Set("Content-Type", "text/plain")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="task-%d.log"`, taskID))

	// Send the file
	return c.SendFile(logFilePath)
}
