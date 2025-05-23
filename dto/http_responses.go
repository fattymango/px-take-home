package dto

import (
	"github.com/gofiber/fiber/v2"
)

func NewSuccessResponse(ctx *fiber.Ctx, data interface{}) error {
	resp := &BaseResponse{
		Success: true,
		Code:    200,
		Data:    data,
		Error:   "",
		Message: "",
	}
	return ctx.Status(200).JSON(resp)
}

func NewSuccessCreatedResponse(ctx *fiber.Ctx, data interface{}) error {
	resp := &BaseResponse{
		Success: true,
		Code:    201,
		Data:    data,
		Error:   "",
		Message: "",
	}
	return ctx.Status(201).JSON(resp)

}

func NewSuccessNoContentResponse(ctx *fiber.Ctx) error {
	resp := &BaseResponse{
		Success: true,
		Code:    204,
		Data:    nil,
		Error:   "",
		Message: "",
	}
	return ctx.Status(204).JSON(resp)
}

func NewUnprocessableEntityResponse(ctx *fiber.Ctx, err string) error {
	resp := &BaseResponse{
		Success: false,
		Code:    422,
		Data:    nil,
		Error:   err,
		Message: "",
	}
	return ctx.Status(422).JSON(resp)
}

func NewInternalServerErrorResponse(ctx *fiber.Ctx, err string) error {
	resp := &BaseResponse{
		Success: false,
		Code:    500,
		Data:    nil,
		Error:   err,
		Message: "",
	}
	return ctx.Status(500).JSON(resp)
}

func NewNotFoundResponse(ctx *fiber.Ctx, err string) error {
	resp := &BaseResponse{
		Success: false,
		Code:    404,
		Data:    nil,
		Error:   err,
		Message: "",
	}

	return ctx.Status(404).JSON(resp)
}

func NewBadRequestResponse(ctx *fiber.Ctx, err string) error {
	resp := &BaseResponse{
		Success: false,
		Code:    400,
		Data:    nil,
		Error:   err,
		Message: "",
	}
	return ctx.Status(400).JSON(resp)
}

func NewUnauthorizedResponse(ctx *fiber.Ctx, err string) error {
	resp := &BaseResponse{
		Success: false,
		Code:    401,
		Data:    nil,
		Error:   err,
		Message: "",
	}
	return ctx.Status(401).JSON(resp)
}

func NewForbiddenResponse(ctx *fiber.Ctx, err string) error {
	resp := &BaseResponse{
		Success: false,
		Code:    403,
		Data:    nil,
		Error:   err,
		Message: "",
	}
	return ctx.Status(403).JSON(resp)
}
