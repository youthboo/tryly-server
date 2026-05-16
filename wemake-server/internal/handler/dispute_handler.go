package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/repository"
	"github.com/yourusername/wemake/internal/service"
)

type DisputeHandler struct {
	service *service.DisputeService
}

func NewDisputeHandler(svc *service.DisputeService) *DisputeHandler {
	return &DisputeHandler{service: svc}
}

// POST /orders/:order_id/disputes
func (h *DisputeHandler) Create(c *fiber.Ctx) error {
	type reqBody struct {
		Reason string `json:"reason" validate:"notblank"`
	}
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	orderID, err := c.ParamsInt("order_id")
	if err != nil || orderID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order_id"})
	}
	var req reqBody
	if err := parseAndValidateBody(c, &req, map[string]string{
		"Reason": "reason is required",
	}); err != nil {
		return err
	}
	item, err := h.service.Create(int64(orderID), userID, req.Reason)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create dispute"})
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

// GET /orders/:order_id/disputes
func (h *DisputeHandler) GetByOrderID(c *fiber.Ctx) error {
	orderID, err := c.ParamsInt("order_id")
	if err != nil || orderID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order_id"})
	}
	item, err := h.service.GetByOrderID(int64(orderID))
	if err != nil {
		if repository.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "dispute not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch dispute"})
	}
	return c.JSON(item)
}

// PATCH /disputes/:dispute_id
func (h *DisputeHandler) PatchStatus(c *fiber.Ctx) error {
	type reqBody struct {
		Status     string  `json:"status"`
		Resolution *string `json:"resolution"`
	}
	disputeID, err := parsePositiveInt64Param(c, "dispute_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid dispute_id"})
	}
	var req reqBody
	if err := requireBody(c, &req); err != nil {
		return err
	}
	if err := h.service.UpdateStatus(disputeID, req.Status, req.Resolution); err != nil {
		if errors.Is(err, service.ErrInvalidDisputeStatus) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		if repository.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "dispute not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update dispute"})
	}
	return c.JSON(fiber.Map{"message": "dispute updated"})
}
