package quotation

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	"github.com/yourusername/wemake/internal/repository"
	quotationservice "github.com/yourusername/wemake/internal/service/quotation"
)

func createQuotationErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		quotationservice.ErrFactorySuspended:       helper.ErrorMessage(fiber.StatusForbidden, "factory account is suspended"),
		quotationservice.ErrFactoryHighlightInvalid: helper.ErrorMessage(fiber.StatusBadRequest, "HIGHLIGHT_TOO_LONG"),
		quotationservice.ErrPaymentTermsInvalid:    helper.ErrorMessage(fiber.StatusBadRequest, "invalid payment terms"),
		quotationservice.ErrInvalidShippingMethod:  helper.ErrorMessage(fiber.StatusBadRequest, "invalid shipping method"),
		quotationservice.ErrInvalidLineItem:        helper.ErrorMessage(fiber.StatusBadRequest, "invalid line item"),
	}
}

func getQuotationErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{}
}

func listHistoryErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{}
}

func patchQuotationErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		quotationservice.ErrQuotationLocked:         helper.ErrorMessage(fiber.StatusConflict, "LOCKED"),
		quotationservice.ErrNotQuotationParty:       helper.ErrorMessage(fiber.StatusForbidden, "not authorized to modify this quotation"),
		quotationservice.ErrFactoryHighlightInvalid: helper.ErrorMessage(fiber.StatusBadRequest, "HIGHLIGHT_TOO_LONG"),
		quotationservice.ErrInvalidShippingMethod:   helper.ErrorMessage(fiber.StatusBadRequest, "invalid shipping method"),
	}
}

func createDetailedErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		quotationservice.ErrFactorySuspended:       helper.ErrorMessage(fiber.StatusForbidden, "factory account is suspended"),
		quotationservice.ErrFactoryHighlightInvalid: helper.ErrorMessage(fiber.StatusBadRequest, "HIGHLIGHT_TOO_LONG"),
	}
}

func canViewErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{}
}

// isNotFoundError checks if an error is a repository not found error
func isNotFoundError(err error) bool {
	return repository.IsNotFoundError(err)
}
