package wallet

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/domainutil"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	walletrepo "github.com/yourusername/wemake/internal/repository/wallet"
	walletservice "github.com/yourusername/wemake/internal/service/wallet"
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
		return helper.JSONInternal(c, "failed to create transaction")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *TransactionHandler) ListTransactions(c *fiber.Ctx) error {
	query := helper.QueryParams(c)
	filters := walletrepo.TransactionFilters{}

	filters.WalletID = query.OptionalPositiveInt64("wallet_id")
	filters.OrderID = query.OptionalPositiveInt64("order_id")
	if err := query.Err(); err != nil {
		return err
	}
	if raw := query.String("type"); raw != "" {
		filters.Type = &raw
	}
	if raw := query.String("status"); raw != "" {
		filters.Status = &raw
	}

	items, err := h.service.List(filters)
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch transactions")
	}
	return c.JSON(items)
}

func (h *TransactionHandler) PatchTransactionStatus(c *fiber.Ctx) error {
	txID := helper.ParamString(c, "tx_id")
	if txID == "" {
		return helper.BadRequestError(c, "invalid tx_id")
	}

	var req dto.PatchTransactionStatusRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}

	status := domainutil.NormalizeStatus(req.Status)
	v := domain.NewValidationCollector()
	v.AddIf(!domainutil.StatusIn(status, "ST", "PT", "RJ"), "status", "must be ST, PT or RJ")
	if err := helper.ValidateRequest(c, v); err != nil {
		return err
	}

	if err := h.service.PatchStatus(txID, status); err != nil {
		return helper.JSONInternal(c, "failed to update transaction status")
	}
	return c.JSON(fiber.Map{"message": "transaction status updated"})
}
