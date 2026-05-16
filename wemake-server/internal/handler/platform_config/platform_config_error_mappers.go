package platformconfig

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	platformservice "github.com/yourusername/wemake/internal/service/platform_config"
)

func createConfigErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		platformservice.ErrPlatformConfigValidation: helper.ErrorMessage(fiber.StatusBadRequest, "invalid platform config payload"),
	}
}

func updateConfigErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		platformservice.ErrPlatformConfigValidation: helper.ErrorMessage(fiber.StatusBadRequest, "invalid platform config payload"),
		platformservice.ErrPlatformConfigNotFound:   helper.ErrorMessage(fiber.StatusNotFound, "platform config not found"),
	}
}

func deleteConfigErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		platformservice.ErrPlatformConfigValidation: helper.ErrorMessage(fiber.StatusBadRequest, "failed to delete platform config"),
		platformservice.ErrPlatformDefaultDelete:    helper.ErrorMessage(fiber.StatusBadRequest, "failed to delete platform config"),
		platformservice.ErrPlatformConfigNotFound:   helper.ErrorMessage(fiber.StatusNotFound, "platform config not found"),
	}
}
