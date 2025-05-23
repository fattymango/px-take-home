package server

import (
	"github.com/fattymango/px-take-home/dto"
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
	task := &dto.CrtTask{}
	if err := c.BodyParser(task); err != nil {
		return dto.NewBadRequestResponse(c, err.Error())
	}

	if err := s.validator.Struct(task); err != nil {
		return dto.NewBadRequestResponse(c, err.Error())
	}

	err := s.TaskManager.CreateTask(task.ToTask())
	if err != nil {
		return dto.NewInternalServerErrorResponse(c, err.Error())
	}

	return dto.NewSuccessResponse(c, task)
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
	tasks, err := s.TaskManager.GetAllTasks()
	if err != nil {
		return dto.NewInternalServerErrorResponse(c, err.Error())
	}

	viewTasks := make([]*dto.ViewTask, len(tasks))
	for i, task := range tasks {
		viewTasks[i] = dto.ToViewTask(task)
	}

	return dto.NewSuccessResponse(c, viewTasks)

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
