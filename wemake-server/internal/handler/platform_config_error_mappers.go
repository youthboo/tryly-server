package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	"github.com/yourusername/wemake/internal/service"
)

func createConfigErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		service.ErrPlatformConfigValidation: helper.ErrorMessage(fiber.StatusBadRequest, "invalid platform config payload"),
	}
}

func updateConfigErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		service.ErrPlatformConfigValidation: helper.ErrorMessage(fiber.StatusBadRequest, "invalid platform config payload"),
		service.ErrPlatformConfigNotFound:   helper.ErrorMessage(fiber.StatusNotFound, "platform config not found"),
	}
}

func deleteConfigErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		service.ErrPlatformConfigValidation: helper.ErrorMessage(fiber.StatusBadRequest, "failed to delete platform config"),
		service.ErrPlatformDefaultDelete:    helper.ErrorMessage(fiber.StatusBadRequest, "failed to delete platform config"),
		service.ErrPlatformConfigNotFound:   helper.ErrorMessage(fiber.StatusNotFound, "platform config not found"),
	}
}
