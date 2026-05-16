package wallet

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
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
	type reqBody struct {
		Amount        float64 `json:"amount" validate:"gt=0"`
		BankAccountNo string  `json:"bank_account_no" validate:"notblank"`
		BankName      string  `json:"bank_name" validate:"notblank"`
		AccountName   string  `json:"account_name" validate:"notblank"`
	}
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req reqBody
	if err := helper.ParseAndValidateBody(c, &req, map[string]string{
		"Amount":        "amount must be greater than 0",
		"BankAccountNo": "bank_account_no, bank_name, and account_name are required",
		"BankName":      "bank_account_no, bank_name, and account_name are required",
		"AccountName":   "bank_account_no, bank_name, and account_name are required",
	}); err != nil {
		return err
	}
	item, err := h.service.Create(userID, req.Amount, req.BankAccountNo, req.BankName, req.AccountName)
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
	type reqBody struct {
		Status string  `json:"status"`
		Note   *string `json:"note"`
	}
	requestID, err := helper.ParsePositiveInt64Param(c, "request_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request_id"})
	}
	var req reqBody
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	if err := h.service.UpdateStatus(requestID, req.Status, req.Note); err != nil {
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
