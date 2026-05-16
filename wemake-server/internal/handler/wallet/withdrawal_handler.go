package wallet

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/repository"
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
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req reqBody
	if err := parseAndValidateBody(c, &req, map[string]string{
		"Amount":        "amount must be greater than 0",
		"BankAccountNo": "bank_account_no, bank_name, and account_name are required",
		"BankName":      "bank_account_no, bank_name, and account_name are required",
		"AccountName":   "bank_account_no, bank_name, and account_name are required",
	}); err != nil {
		return err
	}
	item, err := h.service.Create(userID, req.Amount, req.BankAccountNo, req.BankName, req.AccountName)
	if err != nil {
		if errors.Is(err, walletservice.ErrInsufficientFunds) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		if repository.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wallet not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create withdrawal request"})
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

// GET /wallets/withdraw
func (h *WithdrawalHandler) List(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
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
	requestID, err := parsePositiveInt64Param(c, "request_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request_id"})
	}
	var req reqBody
	if err := requireBody(c, &req); err != nil {
		return err
	}
	if err := h.service.UpdateStatus(requestID, req.Status, req.Note); err != nil {
		if errors.Is(err, walletservice.ErrInvalidWithdrawalStatus) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		if repository.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "withdrawal request not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update withdrawal status"})
	}
	return c.JSON(fiber.Map{"message": "withdrawal status updated"})
}
