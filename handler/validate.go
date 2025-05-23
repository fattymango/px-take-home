package server

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func (s *Server) registerValidator() {

	s.validator.RegisterValidation("not_malformed_command", s.validateNotMalformedCommand)
}

func (s *Server) validateNotMalformedCommand(fl validator.FieldLevel) bool {
	command := fl.Field().String()
	return strings.Contains(command, " ")
}
