package wallet

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	walletservice "github.com/yourusername/wemake/internal/service/wallet"
)

type WalletHandler struct {
	service *walletservice.WalletService
}

func NewWalletHandler(service *walletservice.WalletService) *WalletHandler {
	return &WalletHandler{service: service}
}

func (h *WalletHandler) GetMyWallet(c *fiber.Ctx) error {
	userID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}
	wallet, err := h.service.GetByUserID(userID)
	if err != nil {
		return helper.MapServiceError(c, err, walletGetFallback, walletGetResponses)
	}
	return c.JSON(wallet)
}

func (h *WalletHandler) ListMyTransactions(c *fiber.Ctx) error {
	userID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}
	var orderID *int64
	if raw := c.Query("order_id"); raw != "" {
		val, parseErr := helper.ParsePositiveInt64Value(raw, "order_id")
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order_id"})
		}
		orderID = &val
	}
	var txType *string
	if raw := c.Query("type"); raw != "" {
		txType = &raw
	}
	var status *string
	if raw := c.Query("status"); raw != "" {
		status = &raw
	}
	items, err := h.service.ListTransactionsByUserID(userID, orderID, txType, status)
	if err != nil {
		return helper.MapServiceError(c, err, walletListTransactionsFallback, walletListTransactionsResponses)
	}
	return c.JSON(items)
}

var walletGetFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to fetch wallet")

var walletGetResponses = map[error]helper.ErrorResponse{
	sql.ErrNoRows: helper.ErrorMessage(fiber.StatusNotFound, "wallet not found"),
}

var walletListTransactionsFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to fetch transactions")

var walletListTransactionsResponses = map[error]helper.ErrorResponse{
	sql.ErrNoRows: helper.ErrorMessage(fiber.StatusNotFound, "wallet not found"),
}
