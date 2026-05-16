package conversation

import (
	"errors"
	"github.com/yourusername/wemake/internal/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
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
	convID, err := c.ParamsInt("conv_id")
	if err != nil {
		return helper.BadRequest(c, "invalid conv_id")
	}
	item, err := h.service.GetByID(int64(convID), userID)
	if err != nil {
		if errors.Is(err, conversationservice.ErrConversationForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "conversation not found"})
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
	convID, err := c.ParamsInt("conv_id")
	if err != nil {
		return helper.BadRequest(c, "invalid conv_id")
	}
	if err := h.service.MarkAsRead(int64(convID), userID); err != nil {
		if errors.Is(err, conversationservice.ErrConversationForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		if errors.Is(err, conversationservice.ErrConversationNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "conversation not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to mark conversation as read"})
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
	convID, err := c.ParamsInt("conv_id")
	if err != nil || convID <= 0 {
		return helper.BadRequest(c, "invalid conv_id")
	}
	var req struct {
		RFQID int64 `json:"rfq_id"`
	}
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	if req.RFQID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "rfq_id is required"})
	}
	msg, rfq, err := h.service.ShareRFQ(int64(convID), userID, req.RFQID)
	if err != nil {
		switch {
		case errors.Is(err, conversationservice.ErrShareRFQForbidden):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		case errors.Is(err, conversationservice.ErrShareRFQClosed):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "rfq cannot be shared"})
		case errors.Is(err, conversationservice.ErrShareRFQInvalid):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rfq or conversation not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to share rfq"})
		}
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": msg,
		"rfq":     rfq,
	})
}
