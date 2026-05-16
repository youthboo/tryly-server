package payment

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	paymentservice "github.com/yourusername/wemake/internal/service/payment"
)

type OrderPaymentHandler struct {
	service *paymentservice.OrderPaymentService
}

func NewOrderPaymentHandler(service *paymentservice.OrderPaymentService) *OrderPaymentHandler {
	return &OrderPaymentHandler{service: service}
}

func (h *OrderPaymentHandler) PayDeposit(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	orderID, err := c.ParamsInt("order_id")
	if err != nil || orderID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error_code": "INVALID_ORDER_ID", "message": "invalid order_id"})
	}

	var req dto.PayDepositRequest
	if err := helper.ParseBody(c, &req, "invalid request payload"); err != nil {
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
		return helper.MapServiceErrorFunc(c, rule, orderPaymentFallback, orderPaymentRuleResponses)
	}
	return helper.MapServiceError(c, err, orderPaymentFallback, orderPaymentResponses)
}

var orderPaymentFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to process payment")

var orderPaymentResponses = map[error]helper.ErrorResponse{
	sql.ErrNoRows: helper.ErrorMessage(fiber.StatusNotFound, "order not found"),
}

var orderPaymentRuleResponses = map[error]helper.ErrorResponseBuilder{
	paymentservice.ErrPaymentAmountMismatch: paymentRuleBody(fiber.StatusBadRequest, fiber.Map{
		"error_code": "AMOUNT_MISMATCH",
		"message":    "amount does not match order total amount",
	}),
	paymentservice.ErrPaymentInsufficientWallet: func(err error) helper.ErrorResponse {
		rule, _ := paymentservice.AsPaymentRuleError(err)
		shortfall := 0.0
		if rule != nil {
			shortfall = rule.Shortfall
		}
		return helper.ErrorBody(fiber.StatusBadRequest, fiber.Map{
			"error_code": "INSUFFICIENT_WALLET_BALANCE",
			"message":    "insufficient wallet balance",
			"shortfall":  shortfall,
		})
	},
	paymentservice.ErrPaymentNotOrderOwner: paymentRuleBody(fiber.StatusForbidden, fiber.Map{
		"error": "not order owner",
	}),
	paymentservice.ErrDepositAlreadyPaid: paymentRuleBody(fiber.StatusConflict, fiber.Map{
		"error_code": "DEPOSIT_ALREADY_PAID",
		"message":    "deposit already paid",
	}),
	paymentservice.ErrDepositExpired: paymentRuleBody(fiber.StatusGone, fiber.Map{
		"error_code": "DEPOSIT_EXPIRED",
		"message":    "deposit expired",
	}),
	paymentservice.ErrPaymentFactoryWalletNotFound: paymentRuleBody(fiber.StatusUnprocessableEntity, fiber.Map{
		"error_code": "FACTORY_WALLET_NOT_FOUND",
		"message":    "factory wallet not found",
	}),
	paymentservice.ErrPaymentMethodNotSupported: paymentRuleBody(fiber.StatusNotImplemented, fiber.Map{
		"error_code": "METHOD_NOT_SUPPORTED",
		"message":    "payment method not supported",
	}),
	paymentservice.ErrPaymentTypeNotSupported: paymentRuleBody(fiber.StatusNotImplemented, fiber.Map{
		"error_code": "TYPE_NOT_SUPPORTED",
		"message":    "payment type not supported",
	}),
	paymentservice.ErrPaymentIdempotencyKeyRequired: paymentRuleBody(fiber.StatusBadRequest, fiber.Map{
		"error_code": "IDEMPOTENCY_KEY_REQUIRED",
		"message":    "idempotency_key is required",
	}),
}

func paymentRuleBody(status int, body fiber.Map) helper.ErrorResponseBuilder {
	return func(error) helper.ErrorResponse {
		return helper.ErrorBody(status, body)
	}
}
