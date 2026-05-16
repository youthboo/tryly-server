package wallet

import (
	"database/sql"

	"github.com/yourusername/wemake/internal/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/dto"
	walletservice "github.com/yourusername/wemake/internal/service/wallet"
	"github.com/yourusername/wemake/internal/domainutil"
)

type SettlementHandler struct {
	service *walletservice.SettlementService
}

func NewSettlementHandler(svc *walletservice.SettlementService) *SettlementHandler {
	return &SettlementHandler{service: svc}
}

// GET /settlements
func (h *SettlementHandler) List(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	items, err := h.service.ListByFactoryID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch settlements"})
	}
	return c.JSON(items)
}

// GET /settlements/:settlement_id
func (h *SettlementHandler) GetByID(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	settlementID, err := helper.ParsePositiveInt64Param(c, "settlement_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid settlement_id"})
	}
	item, err := h.service.GetByID(settlementID, userID)
	if err != nil {
		return helper.MapServiceError(c, err, settlementGetFallback, settlementGetResponses)
	}
	return c.JSON(item)
}

// POST /settlements — create a settlement record (factory or system initiated)
func (h *SettlementHandler) Create(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req dto.CreateSettlementRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	item, err := h.service.Create(userID, nil, req.Amount, req.Notes)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create settlement"})
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

// PATCH /settlements/:settlement_id/status
func (h *SettlementHandler) PatchStatus(c *fiber.Ctx) error {
	settlementID, err := helper.ParsePositiveInt64Param(c, "settlement_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid settlement_id"})
	}
	var req dto.PatchSettlementStatusRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	status := domainutil.NormalizeStatus(req.Status)
	if status != "PR" && status != "CP" && status != "FL" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be PR, CP, or FL"})
	}
	if err := h.service.UpdateStatus(settlementID, status); err != nil {
		return helper.MapServiceError(c, err, settlementPatchStatusFallback, settlementPatchStatusResponses)
	}
	return c.JSON(fiber.Map{"message": "settlement status updated"})
}

var settlementGetFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to fetch settlement")

var settlementGetResponses = map[error]helper.ErrorResponse{
	sql.ErrNoRows: helper.ErrorMessage(fiber.StatusNotFound, "settlement not found"),
}

var settlementPatchStatusFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to update settlement status")

var settlementPatchStatusResponses = map[error]helper.ErrorResponse{
	sql.ErrNoRows: helper.ErrorMessage(fiber.StatusNotFound, "settlement not found"),
}
