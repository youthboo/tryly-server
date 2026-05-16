package wallet

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	walletrepo "github.com/yourusername/wemake/internal/repository/wallet"
	walletservice "github.com/yourusername/wemake/internal/service/wallet"
	"github.com/yourusername/wemake/internal/domainutil"
)

type TransactionHandler struct {
	service *walletservice.TransactionService
}

func NewTransactionHandler(service *walletservice.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func (h *TransactionHandler) CreateTransaction(c *fiber.Ctx) error {
	var req dto.CreateTransactionRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}

	item := &domain.Transaction{
		WalletID: 0,
		OrderID:  nil,
		Type:     req.Type,
		Amount:   req.Amount,
		Status:   "ST",
	}
	if err := h.service.Create(item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create transaction"})
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *TransactionHandler) ListTransactions(c *fiber.Ctx) error {
	filters := walletrepo.TransactionFilters{}

	if raw := c.Query("wallet_id"); raw != "" {
		val, err := helper.ParsePositiveInt64Value(raw, "wallet_id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid wallet_id"})
		}
		filters.WalletID = &val
	}
	if raw := c.Query("order_id"); raw != "" {
		val, err := helper.ParsePositiveInt64Value(raw, "order_id")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order_id"})
		}
		filters.OrderID = &val
	}
	if raw := c.Query("type"); raw != "" {
		filters.Type = &raw
	}
	if raw := c.Query("status"); raw != "" {
		filters.Status = &raw
	}

	items, err := h.service.List(filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch transactions"})
	}
	return c.JSON(items)
}

func (h *TransactionHandler) PatchTransactionStatus(c *fiber.Ctx) error {
	txID := c.Params("tx_id")
	if strings.TrimSpace(txID) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tx_id"})
	}

	var req dto.PatchTransactionStatusRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}

	status := domainutil.NormalizeStatus(req.Status)
	if status != "ST" && status != "PT" && status != "RJ" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be ST, PT or RJ"})
	}

	if err := h.service.PatchStatus(txID, status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update transaction status"})
	}
	return c.JSON(fiber.Map{"message": "transaction status updated"})
}
