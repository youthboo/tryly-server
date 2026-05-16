package wallet

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	walletservice "github.com/yourusername/wemake/internal/service/wallet"
)

type WithdrawalHandler struct {
	service *walletservice.WithdrawalService
}

func NewWithdrawalHandler(svc *walletservice.WithdrawalService) *WithdrawalHandler {
	return &WithdrawalHandler{service: svc}
}

// POST /wallets/withdraw
func (h *WithdrawalHandler) Create(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req dto.WithdrawalRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	bankAccountNo := ""
	if req.AccountNumber != nil {
		bankAccountNo = *req.AccountNumber
	}
	bankName := ""
	if req.BankName != nil {
		bankName = *req.BankName
	}
	accountName := ""
	if req.AccountHolderName != nil {
		accountName = *req.AccountHolderName
	}
	item, err := h.service.Create(userID, req.Amount, bankAccountNo, bankName, accountName)
	if err != nil {
		return helper.MapServiceError(c, err, withdrawalCreateFallback, withdrawalCreateResponses)
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

// GET /wallets/withdraw
func (h *WithdrawalHandler) List(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	items, err := h.service.ListByFactoryID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch withdrawal requests"})
	}
	return c.JSON(items)
}

// PATCH /wallets/withdraw/:request_id/status
func (h *WithdrawalHandler) PatchStatus(c *fiber.Ctx) error {
	requestID, err := helper.ParsePositiveInt64Param(c, "request_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request_id"})
	}
	var req dto.PatchWithdrawalStatusRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	if err := h.service.UpdateStatus(requestID, req.Status, req.Comments); err != nil {
		return helper.MapServiceError(c, err, withdrawalPatchStatusFallback, withdrawalPatchStatusResponses)
	}
	return c.JSON(fiber.Map{"message": "withdrawal status updated"})
}

var withdrawalCreateFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to create withdrawal request")

var withdrawalCreateResponses = map[error]helper.ErrorResponse{
	walletservice.ErrInsufficientFunds: helper.ErrorMessage(fiber.StatusBadRequest, walletservice.ErrInsufficientFunds.Error()),
	sql.ErrNoRows:                      helper.ErrorMessage(fiber.StatusNotFound, "wallet not found"),
}

var withdrawalPatchStatusFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to update withdrawal status")

var withdrawalPatchStatusResponses = map[error]helper.ErrorResponse{
	walletservice.ErrInvalidWithdrawalStatus: helper.ErrorMessage(fiber.StatusBadRequest, walletservice.ErrInvalidWithdrawalStatus.Error()),
	sql.ErrNoRows:                            helper.ErrorMessage(fiber.StatusNotFound, "withdrawal request not found"),
}
