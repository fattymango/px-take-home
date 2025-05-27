package server

import (
	"github.com/fattymango/px-take-home/model"
	"github.com/go-playground/validator/v10"
)

func (s *Server) registerValidator() {
	s.validator.RegisterValidation("task_command_enum", s.validateTaskCommandEnum)
}

func (s *Server) validateTaskCommandEnum(fl validator.FieldLevel) bool {
	command := fl.Field().Uint()
	if _, ok := model.TaskCommand_name[model.TaskCommand(command)]; !ok {
		return false
	}
	return true
}
