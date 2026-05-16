package message

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	messageservice "github.com/yourusername/wemake/internal/service/message"
)

func createMessageErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		messageservice.ErrInvalidReferenceType:        helper.ErrorMessage(fiber.StatusBadRequest, "reference_type must be one of RQ, OD, PD, PM, ID"),
		messageservice.ErrReferencePairRequired:       helper.ErrorMessage(fiber.StatusBadRequest, "reference_type and reference_id must be provided together"),
		messageservice.ErrReferenceNotFound:           helper.ErrorMessage(fiber.StatusBadRequest, "reference_id not found for the given reference_type"),
		messageservice.ErrSenderReceiverSame:          helper.ErrorMessage(fiber.StatusBadRequest, "receiver_id must differ from sender_id"),
		messageservice.ErrConversationNotAccessible:   helper.ErrorMessage(fiber.StatusBadRequest, "conv_id must be a conversation the sender belongs to"),
		messageservice.ErrConversationReceiverMismatch: helper.ErrorMessage(fiber.StatusBadRequest, "receiver_id must match the other participant in conv_id"),
		messageservice.ErrInvalidMessageType:          helper.ErrorMessage(fiber.StatusBadRequest, "message_type must be one of TX, QT, IM, BQ"),
		messageservice.ErrQuoteDataRequired:           helper.ErrorMessage(fiber.StatusBadRequest, "quote_data is required when message_type is QT"),
	}
}

func listByReferenceErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		messageservice.ErrInvalidReferenceType: helper.ErrorMessage(fiber.StatusBadRequest, "reference_type must be one of RQ, OD, PD, PM, ID"),
	}
}
