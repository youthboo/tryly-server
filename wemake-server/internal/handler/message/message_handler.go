package message

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/dto"
	handlerregistry "github.com/yourusername/wemake/internal/handler/errorregistry"
	"github.com/yourusername/wemake/internal/helper"
	messageservice "github.com/yourusername/wemake/internal/service/message"
)

type MessageHandler struct {
	service *messageservice.MessageService
}

func NewMessageHandler(service *messageservice.MessageService) *MessageHandler {
	return &MessageHandler{service: service}
}

func (h *MessageHandler) CreateMessage(c *fiber.Ctx) error {
	senderID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}
	var req dto.CreateMessageRequest
	if err := helper.ParseJSONBody(c, &req, "invalid request payload"); err != nil {
		return err
	}
	if err := helper.ValidatePointerInt64(req.ReceiverID, "receiver_id"); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := helper.ValidatePointerString(req.Content, "content"); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	referenceType := helper.DereferenceString(req.ReferenceType, "")
	attachmentURL := helper.DereferenceString(req.AttachmentURL, "")
	messageType := helper.DereferenceString(req.MessageType, "")

	item := &domain.Message{
		ReferenceType: referenceType,
		ReferenceID:   valueOrZero(req.ReferenceID),
		SenderID:      senderID,
		ReceiverID:    *req.ReceiverID,
		Content:       *req.Content,
		AttachmentURL: attachmentURL,
		ConvID:        req.ConvID,
		MessageType:   messageType,
		QuoteData:     req.QuoteData,
		IsRead:        false,
	}
	if err := h.service.Create(item); err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to create message"), handlerregistry.CreateMessageErrorMap())
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *MessageHandler) ListMessages(c *fiber.Ctx) error {
	userID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}

	convID := c.QueryInt("conv_id", 0)
	if convID > 0 {
		items, err := h.service.ListByConvID(int64(convID))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch messages by conv_id"})
		}
		return c.JSON(items)
	}

	referenceType := c.Query("reference_type")
	referenceIDRaw := c.Query("reference_id")
	if strings.TrimSpace(referenceType) == "" || strings.TrimSpace(referenceIDRaw) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "reference_type and reference_id (or conv_id) are required"})
	}
	referenceID, err := helper.ParseRequiredPositiveInt64Query(c, "reference_id")
	if err != nil {
		return helper.BadRequest(c, "reference_id must be a positive integer")
	}
	items, err := h.service.ListByReference(referenceType, referenceID, userID)
	if err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to fetch messages"), handlerregistry.ListByReferenceErrorMap())
	}
	return c.JSON(items)
}

func (h *MessageHandler) ListThreads(c *fiber.Ctx) error {
	userID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}
	items, err := h.service.ListThreads(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch threads"})
	}
	return c.JSON(items)
}

func valueOrZero(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}
