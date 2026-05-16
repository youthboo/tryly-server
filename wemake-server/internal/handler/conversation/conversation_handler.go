package conversation

import (
	"github.com/yourusername/wemake/internal/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/dto"
	conversationservice "github.com/yourusername/wemake/internal/service/conversation"
)

type ConversationHandler struct {
	service *conversationservice.ConversationService
}

func NewConversationHandler(service *conversationservice.ConversationService) *ConversationHandler {
	return &ConversationHandler{service: service}
}

func (h *ConversationHandler) List(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.Unauthorized(c)
	}
	items, err := h.service.ListByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch conversations"})
	}
	return c.JSON(items)
}

func (h *ConversationHandler) Get(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.Unauthorized(c)
	}
	convID, err := helper.RequireInt64Param(c, "conv_id")
	if err != nil {
		return err
	}
	item, err := h.service.GetByID(int64(convID), userID)
	if err != nil {
		return helper.MapServiceError(c, err, conversationGetFallback, conversationGetResponses)
	}
	return c.JSON(item)
}

func (h *ConversationHandler) Create(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.Unauthorized(c)
	}
	var req domain.Conversation
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	if req.CustomerID <= 0 || req.FactoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "customer_id and factory_id are required"})
	}
	// Allow both customer (CT) and factory (FT) to initiate a conversation room.
	// The caller must be one of the two parties — this is the security boundary.
	if req.CustomerID != userID && req.FactoryID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
	if err := h.service.Create(&req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create conversation"})
	}
	item, err := h.service.GetByID(req.ConvID, userID)
	if err != nil {
		return c.Status(fiber.StatusCreated).JSON(req)
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *ConversationHandler) InquireShowcase(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.Unauthorized(c)
	}
	showcaseID, err := helper.ParsePositiveInt64Param(c, "showcase_id")
	if err != nil {
		return helper.BadRequest(c, "invalid showcase_id")
	}
	if role := helper.OptionalRoleFromContext(c); role != "" && role != domain.RoleCustomer {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "buyer role required"})
	}
	item, err := h.service.CreateFromShowcase(showcaseID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create conversation"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"conv_id": item.ConvID})
}

func (h *ConversationHandler) MarkAsRead(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.Unauthorized(c)
	}
	convID, err := helper.RequireInt64Param(c, "conv_id")
	if err != nil {
		return err
	}
	if err := h.service.MarkAsRead(int64(convID), userID); err != nil {
		return helper.MapServiceError(c, err, conversationMarkAsReadFallback, conversationMarkAsReadResponses)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ConversationHandler) ShareRFQ(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.Unauthorized(c)
	}
	if role := helper.OptionalRoleFromContext(c); role != "" && role != domain.RoleCustomer {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "buyer role required"})
	}
	convID, err := helper.RequireInt64Param(c, "conv_id")
		if err != nil {
			return err
		}
	var req dto.ShareRFQRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	if req.RFQID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "rfq_id is required"})
	}
	msg, rfq, err := h.service.ShareRFQ(int64(convID), userID, req.RFQID)
	if err != nil {
		return helper.MapServiceError(c, err, conversationShareRFQFallback, conversationShareRFQResponses)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": msg,
		"rfq":     rfq,
	})
}

var conversationGetFallback = helper.ErrorMessage(fiber.StatusNotFound, "conversation not found")

var conversationGetResponses = map[error]helper.ErrorResponse{
	conversationservice.ErrConversationForbidden: helper.ErrorMessage(fiber.StatusForbidden, "forbidden"),
}

var conversationMarkAsReadFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to mark conversation as read")

var conversationMarkAsReadResponses = map[error]helper.ErrorResponse{
	conversationservice.ErrConversationForbidden: helper.ErrorMessage(fiber.StatusForbidden, "forbidden"),
	conversationservice.ErrConversationNotFound:  helper.ErrorMessage(fiber.StatusNotFound, "conversation not found"),
}

var conversationShareRFQFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to share rfq")

var conversationShareRFQResponses = map[error]helper.ErrorResponse{
	conversationservice.ErrShareRFQForbidden: helper.ErrorMessage(fiber.StatusForbidden, "forbidden"),
	conversationservice.ErrShareRFQClosed:    helper.ErrorMessage(fiber.StatusUnprocessableEntity, "rfq cannot be shared"),
	conversationservice.ErrShareRFQInvalid:   helper.ErrorMessage(fiber.StatusNotFound, "rfq or conversation not found"),
}
