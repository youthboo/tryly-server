package order

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	orderservice "github.com/yourusername/wemake/internal/service/order"
)

func createOrderErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		orderservice.ErrQuotationRejected:          helper.ErrorMessage(fiber.StatusBadRequest, "quotation was rejected"),
		orderservice.ErrQuotationInvalidState:      helper.ErrorMessage(fiber.StatusBadRequest, "quotation has invalid state"),
		orderservice.ErrInsufficientGoodFund:       helper.ErrorMessage(fiber.StatusBadRequest, "insufficient good fund balance"),
		orderservice.ErrOrderAlreadyExistsForQuote: helper.ErrorMessage(fiber.StatusConflict, "order already exists for this quotation"),
		orderservice.ErrInvalidOrderTotal:          helper.ErrorMessage(fiber.StatusBadRequest, "order total must be non-negative"),
	}
}

func bulkCheckoutErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		orderservice.ErrRFQLocked:                  helper.ErrorMessage(fiber.StatusLocked, "rfq is locked"),
		orderservice.ErrQuotationInvalidState:      helper.ErrorMessage(fiber.StatusConflict, "quotation has invalid state"),
		orderservice.ErrSelfTransaction:            helper.ErrorMessage(fiber.StatusBadRequest, "cannot transact with self"),
		orderservice.ErrInvalidQuotationSet:        helper.ErrorMessage(fiber.StatusBadRequest, "invalid quotation set"),
		orderservice.ErrPaymentTypeInvalid:         helper.ErrorMessage(fiber.StatusBadRequest, "invalid payment type"),
		orderservice.ErrNotQuotationParty:          helper.ErrorMessage(fiber.StatusForbidden, "not authorized for this quotation"),
		orderservice.ErrOrderAlreadyExistsForQuote: helper.ErrorMessage(fiber.StatusConflict, "order already exists for this quotation"),
		orderservice.ErrInvalidOrderTotal:          helper.ErrorMessage(fiber.StatusBadRequest, "order total must be non-negative"),
		orderservice.ErrNoOrdersCreated:            helper.ErrorMessage(fiber.StatusBadRequest, "no valid orders created from quotations"),
	}
}

func createPaymentErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		orderservice.ErrDepositAlreadyPaid:    helper.ErrorBody(fiber.StatusUnprocessableEntity, fiber.Map{"error": fiber.Map{"code": "DEPOSIT_ALREADY_PAID", "message": "deposit already recorded"}}),
		orderservice.ErrDepositExpired:        helper.ErrorBody(fiber.StatusGone, fiber.Map{"error": fiber.Map{"code": "DEPOSIT_EXPIRED", "message": "deposit payment window has expired"}}),
		orderservice.ErrPaymentTypeInvalid:    helper.ErrorMessage(fiber.StatusBadRequest, "invalid payment type"),
		orderservice.ErrPaymentAmountMismatch: helper.ErrorMessage(fiber.StatusBadRequest, "payment amount does not match"),
		orderservice.ErrPaymentAlreadyExists:  helper.ErrorMessage(fiber.StatusBadRequest, "payment already exists"),
	}
}

func verifyPaymentErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		orderservice.ErrDepositAlreadyPaid:    helper.ErrorBody(fiber.StatusUnprocessableEntity, fiber.Map{"error": fiber.Map{"code": "DEPOSIT_ALREADY_PAID", "message": "deposit already recorded"}}),
		orderservice.ErrDepositExpired:        helper.ErrorBody(fiber.StatusGone, fiber.Map{"error": fiber.Map{"code": "DEPOSIT_EXPIRED", "message": "deposit payment window has expired"}}),
		orderservice.ErrPaymentTypeInvalid:    helper.ErrorMessage(fiber.StatusBadRequest, "invalid payment type"),
		orderservice.ErrPaymentAmountMismatch: helper.ErrorMessage(fiber.StatusBadRequest, "payment amount does not match"),
		orderservice.ErrPaymentStateInvalid:   helper.ErrorMessage(fiber.StatusBadRequest, "invalid payment state"),
		orderservice.ErrInsufficientGoodFund:  helper.ErrorMessage(fiber.StatusBadRequest, "insufficient good fund balance"),
	}
}

func confirmReceiptErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		orderservice.ErrConfirmReceiptInvalidStatus: helper.ErrorMessage(fiber.StatusConflict, "order cannot be marked as received at this stage"),
		orderservice.ErrConfirmReceiptNotAllowed:    helper.ErrorMessage(fiber.StatusUnprocessableEntity, "confirm receipt not allowed for this order"),
	}
}
