package wallet

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	walletservice "github.com/yourusername/wemake/internal/service/wallet"
)

type DisputeHandler struct {
	service *walletservice.DisputeService
}

func NewDisputeHandler(svc *walletservice.DisputeService) *DisputeHandler {
	return &DisputeHandler{service: svc}
}

// POST /orders/:order_id/disputes
func (h *DisputeHandler) Create(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	orderID, err := helper.RequireInt64Param(c, "order_id")
	if err != nil {
		return err
	}
	var req dto.CreateDisputeRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	item, err := h.service.Create(int64(orderID), userID, req.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create dispute"})
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

// GET /orders/:order_id/disputes
func (h *DisputeHandler) GetByOrderID(c *fiber.Ctx) error {
	orderID, err := helper.RequireInt64Param(c, "order_id")
	if err != nil {
		return err
	}
	item, err := h.service.GetByOrderID(int64(orderID))
	if err != nil {
		return helper.MapServiceError(c, err, disputeGetFallback, disputeGetResponses)
	}
	return c.JSON(item)
}

// PATCH /disputes/:dispute_id
func (h *DisputeHandler) PatchStatus(c *fiber.Ctx) error {
	disputeID, err := helper.ParsePositiveInt64Param(c, "dispute_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid dispute_id"})
	}
	var req dto.PatchDisputeStatusRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	if err := h.service.UpdateStatus(disputeID, req.Status, req.Comments); err != nil {
		return helper.MapServiceError(c, err, disputePatchStatusFallback, disputePatchStatusResponses)
	}
	return c.JSON(fiber.Map{"message": "dispute updated"})
}

var disputeGetFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to fetch dispute")

var disputeGetResponses = map[error]helper.ErrorResponse{
	sql.ErrNoRows: helper.ErrorMessage(fiber.StatusNotFound, "dispute not found"),
}

var disputePatchStatusFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to update dispute")

var disputePatchStatusResponses = map[error]helper.ErrorResponse{
	walletservice.ErrInvalidDisputeStatus: helper.ErrorMessage(fiber.StatusBadRequest, walletservice.ErrInvalidDisputeStatus.Error()),
	sql.ErrNoRows:                         helper.ErrorMessage(fiber.StatusNotFound, "dispute not found"),
}
