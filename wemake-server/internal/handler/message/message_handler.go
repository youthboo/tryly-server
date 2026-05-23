package message

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/dto"
	handlerregistry "github.com/yourusername/wemake/internal/handler/errorregistry"
	"github.com/yourusername/wemake/internal/helper"
	"github.com/yourusername/wemake/internal/logger"
	"github.com/yourusername/wemake/internal/sse"
	messageservice "github.com/yourusername/wemake/internal/service/message"
)

type MessageHandler struct {
	service *messageservice.MessageService
	hub     *sse.Hub
}

func NewMessageHandler(service *messageservice.MessageService, hub *sse.Hub) *MessageHandler {
	return &MessageHandler{service: service, hub: hub}
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
	v := domain.NewValidationCollector()
	v.AddIf(req.ReceiverID == nil || *req.ReceiverID <= 0, "receiver_id", "is required and must be positive")
	v.AddIf(req.Content == nil || helper.DereferenceString(req.Content, "") == "", "content", "is required")
	if err := helper.ValidateRequest(c, v); err != nil {
		return err
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
	pushNewMessageSSE(h.hub, item)
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *MessageHandler) ListMessages(c *fiber.Ctx) error {
	userID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}

	query := helper.QueryParams(c)
	convID := query.OptionalPositiveInt64("conv_id")
	if err := query.Err(); err != nil {
		return err
	}
	if convID != nil {
		items, err := h.service.ListByConvID(*convID)
		if err != nil {
			return helper.JSONInternal(c, "failed to fetch messages by conv_id")
		}
		return c.JSON(items)
	}

	referenceType := query.String("reference_type")
	if referenceType == "" || query.String("reference_id") == "" {
		return helper.BadRequestError(c, "reference_type and reference_id (or conv_id) are required")
	}
	referenceID := query.RequiredPositiveInt64("reference_id")
	if query.Err() != nil {
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
		return helper.JSONInternal(c, "failed to fetch threads")
	}
	return c.JSON(items)
}

// ListMessagesByConvPath handles GET /conversations/:conv_id/messages
// It extracts conv_id from the URL path instead of query params.
func (h *MessageHandler) ListMessagesByConvPath(c *fiber.Ctx) error {
	_, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}
	convID, err := helper.RequireInt64Param(c, "conv_id")
	if err != nil {
		return err
	}
	items, err := h.service.ListByConvID(int64(convID))
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch messages")
	}
	return c.JSON(items)
}

// CreateMessageByConvPath handles POST /conversations/:conv_id/messages
// It injects conv_id from URL path into the request body.
func (h *MessageHandler) CreateMessageByConvPath(c *fiber.Ctx) error {
	senderID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}
	convID, err := helper.RequireInt64Param(c, "conv_id")
	if err != nil {
		return err
	}

	var req dto.CreateMessageRequest
	if err := helper.ParseJSONBody(c, &req, "invalid request payload"); err != nil {
		return err
	}
	// Override conv_id from path
	pathConvID := int64(convID)
	req.ConvID = &pathConvID

	v := domain.NewValidationCollector()
	v.AddIf(req.ReceiverID == nil || *req.ReceiverID <= 0, "receiver_id", "is required and must be positive")
	v.AddIf(req.Content == nil || helper.DereferenceString(req.Content, "") == "", "content", "is required")
	if err := helper.ValidateRequest(c, v); err != nil {
		return err
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
		logger.Error("create message by conv path failed", "sender_id", item.SenderID, "receiver_id", item.ReceiverID, "conv_id", item.ConvID, "err", err)
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to create message"), handlerregistry.CreateMessageErrorMap())
	}
	pushNewMessageSSE(h.hub, item)
	return c.Status(fiber.StatusCreated).JSON(item)
}

func pushNewMessageSSE(hub *sse.Hub, msg *domain.Message) {
	if hub == nil {
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	hub.Push(msg.ReceiverID, "new_message", string(data))
	hub.Push(msg.SenderID, "new_message", string(data))
}

func valueOrZero(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}
