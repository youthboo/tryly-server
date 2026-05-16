package payment

import (
	"database/sql"
	"errors"
	"github.com/yourusername/wemake/internal/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/repository"
	paymentservice "github.com/yourusername/wemake/internal/service/payment"
)

type OrderPaymentHandler struct {
	service *paymentservice.OrderPaymentService
}

func NewOrderPaymentHandler(service *paymentservice.OrderPaymentService) *OrderPaymentHandler {
	return &OrderPaymentHandler{service: service}
}

func (h *OrderPaymentHandler) PayDeposit(c *fiber.Ctx) error {
	type reqBody struct {
		Type           string  `json:"type"`
		Amount         float64 `json:"amount"`
		PaymentMethod  string  `json:"payment_method"`
		IdempotencyKey string  `json:"idempotency_key"`
	}

	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	orderID, err := c.ParamsInt("order_id")
	if err != nil || orderID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error_code": "INVALID_ORDER_ID", "message": "invalid order_id"})
	}

	var req reqBody
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error_code": "INVALID_PAYLOAD", "message": "invalid request payload"})
	}

	out, err := h.service.PayDeposit(paymentservice.OrderPaymentInput{
		OrderID:        int64(orderID),
		UserID:         userID,
		Type:           req.Type,
		Amount:         req.Amount,
		PaymentMethod:  req.PaymentMethod,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		return orderPaymentError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func orderPaymentError(c *fiber.Ctx, err error) error {
	if rule, ok := paymentservice.AsPaymentRuleError(err); ok {
		switch {
		case errors.Is(rule, paymentservice.ErrPaymentAmountMismatch):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error_code": "AMOUNT_MISMATCH", "message": "amount does not match order total amount"})
		case errors.Is(rule, paymentservice.ErrPaymentInsufficientWallet):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error_code": "INSUFFICIENT_WALLET_BALANCE",
				"message":    "insufficient wallet balance",
				"shortfall":  rule.Shortfall,
			})
		case errors.Is(rule, paymentservice.ErrPaymentNotOrderOwner):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not order owner"})
		case errors.Is(rule, paymentservice.ErrDepositAlreadyPaid):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error_code": "DEPOSIT_ALREADY_PAID", "message": "deposit already paid"})
		case errors.Is(rule, paymentservice.ErrDepositExpired):
			return c.Status(fiber.StatusGone).JSON(fiber.Map{"error_code": "DEPOSIT_EXPIRED", "message": "deposit expired"})
		case errors.Is(rule, paymentservice.ErrPaymentFactoryWalletNotFound):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error_code": "FACTORY_WALLET_NOT_FOUND", "message": "factory wallet not found"})
		case errors.Is(rule, paymentservice.ErrPaymentMethodNotSupported):
			return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error_code": "METHOD_NOT_SUPPORTED", "message": "payment method not supported"})
		case errors.Is(rule, paymentservice.ErrPaymentTypeNotSupported):
			return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error_code": "TYPE_NOT_SUPPORTED", "message": "payment type not supported"})
		case errors.Is(rule, paymentservice.ErrPaymentIdempotencyKeyRequired):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error_code": "IDEMPOTENCY_KEY_REQUIRED", "message": "idempotency_key is required"})
		}
	}
	if errors.Is(err, sql.ErrNoRows) || repository.IsNotFoundError(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "order not found"})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to process payment"})
}
