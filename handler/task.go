package server

import (
	"fmt"

	"github.com/fattymango/px-take-home/dto"
	"github.com/fattymango/px-take-home/internal/task"
	"github.com/fattymango/px-take-home/pkg/ctxstore"
	"github.com/gofiber/fiber/v2"
)

// @Tags Task
// @Summary Create task
// @Router /api/v1/tasks [post]
// @Security BearerAuth
// @Description Create task
// @Accept json
// @Produce json
//
//	@Success	200	{object} dto.ViewTask "Success"
//	@Failure	400	{object} dto.BaseResponse	"Bad Request"
//	@Failure	401	{object} dto.BaseResponse	"Unauthorized"
//	@Failure	404	{object} dto.BaseResponse	"Not Found"
//	@Failure	500	{object} dto.BaseResponse	"Internal Server Error"
//
// @Security BearerAuth
// @ID CreateTask
func (s *Server) CreateTask(c *fiber.Ctx) error {
	crt := &dto.CrtTask{}
	if err := c.BodyParser(crt); err != nil {
		return dto.NewBadRequestResponse(c, err.Error())
	}

	if err := s.validator.Struct(crt); err != nil {
		return dto.NewBadRequestResponse(c, err.Error())
	}

	task, err := s.TaskManager.CreateTask(crt.ToTask())
	if err != nil {
		return dto.NewInternalServerErrorResponse(c, fmt.Sprintf("failed to create task: %s", err))
	}

	err = s.TaskManager.QueueTask(task)
	if err != nil {
		return dto.NewInternalServerErrorResponse(c, fmt.Sprintf("failed to queue task: %s", err))
	}

	return dto.NewSuccessResponse(c, dto.ToViewTask(task))
}

// @Tags Task
// @Summary Get all tasks
// @Router /api/v1/tasks [get]
// @Security BearerAuth
// @Description Get all tasks
// @Accept json
// @Produce json
//
//	@Success	200	{object} dto.ViewTask "Success"
//	@Failure	400	{object} dto.BaseResponse	"Bad Request"
//	@Failure	401	{object} dto.BaseResponse	"Unauthorized"
//	@Failure	404	{object} dto.BaseResponse	"Not Found"
//	@Failure	500	{object} dto.BaseResponse	"Internal Server Error"
//
// @Security BearerAuth
// @ID GetAllTasks
func (s *Server) GetAllTasks(c *fiber.Ctx) error {
	offset, limit := ctxstore.GetOffsetLimitQueryFromCtx(c)
	status, err := ctxstore.GetStatusQueryFromCtx(c)
	if err != nil {
		return dto.NewBadRequestResponse(c, err.Error())
	}

	tasks, total, err := s.TaskManager.GetAllTasks(offset, limit, status)
	if err != nil {
		return dto.NewInternalServerErrorResponse(c, err.Error())
	}

	return dto.NewSuccessResponse(c, dto.ToListTasks(tasks, total))
}

// @Tags Task
// @Summary Get task by ID
// @Router /api/v1/tasks/{taskID} [get]
// @Security BearerAuth
// @Description Get task by ID
// @Accept json
// @Produce json
//
//	@Success	200	{object} dto.ViewTask "Success"
//	@Failure	400	{object} dto.BaseResponse	"Bad Request"
//	@Failure	401	{object} dto.BaseResponse	"Unauthorized"
//	@Failure	404	{object} dto.BaseResponse	"Not Found"
//	@Failure	500	{object} dto.BaseResponse	"Internal Server Error"
//
// @Security BearerAuth
// @ID GetTaskByID
func (s *Server) GetTaskByID(c *fiber.Ctx) error {
	taskID, err := ctxstore.GetTaskIDFromCtx(c)
	if err != nil {
		return dto.NewBadRequestResponse(c, err.Error())
	}

	task, err := s.TaskManager.GetTask(taskID)
	if err != nil {
		return dto.NewInternalServerErrorResponse(c, err.Error())
	}

	return dto.NewSuccessResponse(c, dto.ToViewTask(task))
}

// @Tags Task
// @Summary Cancel task
// @Router /api/v1/tasks/{taskID}/cancel [delete]
// @Security BearerAuth
// @Description Cancel task
// @Accept json
// @Produce json
//
//	@Success	200	{object} dto.BaseResponse "Success"
//	@Failure	400	{object} dto.BaseResponse	"Bad Request"
//	@Failure	401	{object} dto.BaseResponse	"Unauthorized"
//	@Failure	404	{object} dto.BaseResponse	"Not Found"
//	@Failure	500	{object} dto.BaseResponse	"Internal Server Error"
//
// @Security BearerAuth
// @ID CancelTask
func (s *Server) CancelTask(c *fiber.Ctx) error {
	taskID, err := ctxstore.GetTaskIDFromCtx(c)
	if err != nil {
		return dto.NewBadRequestResponse(c, err.Error())
	}

	err = s.TaskManager.CancelTask(taskID, task.ReasonCancelledByUser, -1)
	if err != nil {
		return dto.NewInternalServerErrorResponse(c, err.Error())
	}

	return dto.NewSuccessResponse(c, nil)
}
