package boq

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	boqservice "github.com/yourusername/wemake/internal/service/boq"
)

func boqErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		boqservice.ErrBOQNotFound:        helper.ErrorMessage(fiber.StatusNotFound, "BOQ_NOT_FOUND"),
		boqservice.ErrBOQForbidden:       helper.ErrorMessage(fiber.StatusForbidden, "BOQ_FORBIDDEN"),
		boqservice.ErrBOQInvalidItems:    helper.ErrorMessage(fiber.StatusBadRequest, "BOQ_INVALID_ITEMS"),
		boqservice.ErrBOQExpired:         helper.ErrorMessage(fiber.StatusUnprocessableEntity, "BOQ_EXPIRED"),
		boqservice.ErrBOQAlreadyHandled:  helper.ErrorMessage(fiber.StatusConflict, "BOQ_ALREADY_HANDLED"),
		boqservice.ErrBOQInvalidState:    helper.ErrorMessage(fiber.StatusConflict, "BOQ_INVALID_STATE"),
	}
}
