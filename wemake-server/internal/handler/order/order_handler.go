package order

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	"github.com/yourusername/wemake/internal/repository"
	"github.com/yourusername/wemake/internal/service"
	orderservice "github.com/yourusername/wemake/internal/service/order"
)

type OrderHandler struct {
	service *orderservice.OrderService
	auth    *service.AuthService
}

func NewOrderHandler(orderService *orderservice.OrderService, authService *service.AuthService) *OrderHandler {
	return &OrderHandler{service: orderService, auth: authService}
}

func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	var req dto.CreateOrderFromQuoteRequest
	if err := helper.ParseAndValidateBody(c, &req, map[string]string{
		"QuotationID": "quote_id is required",
	}); err != nil {
		return err
	}
	order, err := h.service.CreateFromQuotation(req.QuotationID, userID)
	if err != nil {
		if errors.Is(err, orderservice.ErrQuotationRejected) ||
			errors.Is(err, orderservice.ErrQuotationInvalidState) ||
			errors.Is(err, orderservice.ErrInsufficientGoodFund) ||
			errors.Is(err, orderservice.ErrOrderAlreadyExistsForQuote) {
			return helper.WriteServiceErrorWithNotFound(c, err, "failed to create order", "order not found",
				helper.BadRequestCase(orderservice.ErrQuotationRejected),
				helper.BadRequestCase(orderservice.ErrQuotationInvalidState),
				helper.BadRequestCase(orderservice.ErrInsufficientGoodFund),
				helper.ConflictCase(orderservice.ErrOrderAlreadyExistsForQuote),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "failed to create order",
			"detail": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(order)
}

func (h *OrderHandler) BulkCheckout(c *fiber.Ctx) error {
	type reqBody struct {
		Items          []orderservice.BulkCheckoutItemInput `json:"items"`
		IdempotencyKey string                               `json:"idempotency_key"`
	}
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	rfqID, err := c.ParamsInt("rfq_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
	}
	var req reqBody
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	result, err := h.service.BulkCheckout(orderservice.BulkCheckoutInput{
		RFQID:          int64(rfqID),
		UserID:         userID,
		Items:          req.Items,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		switch {
		case errors.Is(err, orderservice.ErrRFQLocked):
			return c.Status(fiber.StatusLocked).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, orderservice.ErrQuotationInvalidState):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "QUOTATION_NOT_PENDING"})
		case errors.Is(err, orderservice.ErrSelfTransaction):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, orderservice.ErrInvalidQuotationSet), errors.Is(err, orderservice.ErrPaymentTypeInvalid):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, orderservice.ErrNotQuotationParty):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, orderservice.ErrOrderAlreadyExistsForQuote):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to bulk checkout"})
		}
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *OrderHandler) ListOrders(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	if factoryParam := strings.TrimSpace(c.Query("factory_id")); factoryParam != "" {
		if u.Role != domain.RoleFactory {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "factory role required"})
		}
		if !strings.EqualFold(factoryParam, "me") {
			factoryID, parseErr := helper.ParsePositiveInt64Value(factoryParam, "factory_id")
			if parseErr != nil || factoryID <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid factory_id"})
			}
			if factoryID != userID {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "factory_id must match authenticated factory"})
			}
		}
	}
	status := strings.TrimSpace(c.Query("status"))
	var rfqID *int64
	if raw := strings.TrimSpace(c.Query("rfq_id")); raw != "" {
		parsed, parseErr := helper.ParsePositiveInt64Value(raw, "rfq_id")
		if parseErr != nil || parsed <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
		}
		rfqID = &parsed
	}
	items, err := h.service.List(userID, u.Role, status, rfqID, c.Query("request_kind"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch orders"})
	}
	return c.JSON(items)
}

func (h *OrderHandler) GetOrder(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	orderID, err := c.ParamsInt("order_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order_id"})
	}
	detail, err := h.service.GetDetailByID(int64(orderID), userID, u.Role)
	if err != nil {
		return helper.WriteServiceErrorWithNotFound(c, err, "failed to fetch order", "order not found")
	}
	return c.JSON(detail)
}

func (h *OrderHandler) ListActivity(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	orderID, err := c.ParamsInt("order_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order_id"})
	}
	if _, err := h.service.GetByID(int64(orderID), userID, u.Role); err != nil {
		return helper.WriteServiceErrorWithNotFound(c, err, "failed to verify order", "order not found")
	}
	items, err := h.service.ListActivity(int64(orderID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch activity"})
	}
	return c.JSON(items)
}

func (h *OrderHandler) PatchOrderStatus(c *fiber.Ctx) error {
	uid, authErr := helper.UserIDFromHeader(c)
	var actor *int64
	if authErr == nil {
		actor = &uid
	}
	orderID, err := c.ParamsInt("order_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order_id"})
	}
	var req dto.PatchOrderStatusRequest
	if err := helper.ParseAndValidateBody(c, &req, map[string]string{
		"Status": "status must be PP, PR, WF, QC, SH, DL, AC, CP, or CC",
	}); err != nil {
		return err
	}
	status := strings.TrimSpace(strings.ToUpper(req.Status))
	validOrderStatuses := map[string]struct{}{
		"PP": {}, "PR": {}, "WF": {}, "QC": {}, "SH": {}, "DL": {}, "AC": {}, "CP": {}, "CC": {},
	}
	if _, ok := validOrderStatuses[status]; !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be PP, PR, WF, QC, SH, DL, AC, CP, or CC"})
	}
	if err := h.service.UpdateStatus(int64(orderID), status, actor); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update order status"})
	}
	return c.JSON(fiber.Map{"message": "order status updated"})
}

func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	orderID, err := c.ParamsInt("order_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order_id"})
	}
	if err := h.service.Cancel(int64(orderID), userID, u.Role); err != nil {
		return helper.WriteServiceErrorWithNotFound(c, err, "failed to cancel order", "order not found", helper.BadRequestCase(orderservice.ErrOrderCannotBeCancelled))
	}
	return c.JSON(fiber.Map{"message": "order cancelled"})
}

func (h *OrderHandler) MarkShipped(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	if u.Role != domain.RoleFactory {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "factory role required"})
	}
	orderID, err := c.ParamsInt("order_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order_id"})
	}
	var req dto.MarkShippedRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	if err := h.service.MarkShipped(int64(orderID), userID, req.TrackingNo, req.Courier); err != nil {
		return helper.WriteServiceErrorWithNotFound(c, err, "failed to mark order as shipped", "order not found", helper.BadRequestCase(orderservice.ErrShipOrderInvalid))
	}
	item, err := h.service.GetByID(int64(orderID), userID, u.Role)
	if err != nil {
		return c.JSON(fiber.Map{"message": "order marked as shipped"})
	}
	return c.JSON(item)
}

func (h *OrderHandler) CreatePayment(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	orderID, err := c.ParamsInt("order_id")
	if err != nil || orderID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order_id"})
	}
	var req dto.PayDepositRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	item, err := h.service.CreatePayment(int64(orderID), userID, u.Role, req.Type, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, orderservice.ErrDepositAlreadyPaid):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": fiber.Map{"code": "DEPOSIT_ALREADY_PAID", "message": "deposit already recorded"},
			})
		case errors.Is(err, orderservice.ErrDepositExpired):
			return c.Status(fiber.StatusGone).JSON(fiber.Map{
				"error": fiber.Map{"code": "DEPOSIT_EXPIRED", "message": "deposit payment window has expired"},
			})
		case errors.Is(err, orderservice.ErrPaymentTypeInvalid),
			errors.Is(err, orderservice.ErrPaymentAmountMismatch),
			errors.Is(err, orderservice.ErrPaymentAlreadyExists):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case repository.IsNotFoundError(err):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "order not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create payment"})
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *OrderHandler) VerifyPayment(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	orderID, err := c.ParamsInt("order_id")
	if err != nil || orderID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order_id"})
	}
	txID := strings.TrimSpace(c.Params("tx_id"))
	if txID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tx_id"})
	}
	item, err := h.service.VerifyPayment(int64(orderID), userID, u.Role, txID)
	if err != nil {
		switch {
		case errors.Is(err, orderservice.ErrDepositAlreadyPaid):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": fiber.Map{"code": "DEPOSIT_ALREADY_PAID", "message": "deposit already recorded"},
			})
		case errors.Is(err, orderservice.ErrDepositExpired):
			return c.Status(fiber.StatusGone).JSON(fiber.Map{
				"error": fiber.Map{"code": "DEPOSIT_EXPIRED", "message": "deposit payment window has expired"},
			})
		case errors.Is(err, orderservice.ErrPaymentTypeInvalid),
			errors.Is(err, orderservice.ErrPaymentAmountMismatch),
			errors.Is(err, orderservice.ErrPaymentStateInvalid),
			errors.Is(err, orderservice.ErrInsufficientGoodFund):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case repository.IsNotFoundError(err):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "payment not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to verify payment"})
	}
	return c.JSON(item)
}

func (h *OrderHandler) ConfirmReceipt(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	orderID, err := c.ParamsInt("order_id")
	if err != nil || orderID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order_id"})
	}
	var req dto.ConfirmReceiptRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	var receivedAt *time.Time
	if req.ReceivedAt != nil && strings.TrimSpace(*req.ReceivedAt) != "" {
		t, parseErr := helper.ParseOptionalRFC3339Value(req.ReceivedAt, "received_at")
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "received_at must be RFC3339 datetime"})
		}
		receivedAt = t
	}

	result, err := h.service.ConfirmReceipt(int64(orderID), userID, u.Role, orderservice.ConfirmReceiptInput{
		Note:       req.Note,
		ReceivedAt: receivedAt,
	})
	if err != nil {
		return helper.WriteServiceErrorWithNotFound(c, err, "failed to confirm receipt", "order not found",
			helper.ConflictCase(orderservice.ErrConfirmReceiptInvalidStatus),
			helper.UnprocessableCase(orderservice.ErrConfirmReceiptNotAllowed),
		)
	}
	return c.JSON(result)
}

func (h *OrderHandler) GetReviewState(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	orderID, err := c.ParamsInt("order_id")
	if err != nil || orderID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order_id"})
	}
	item, err := h.service.GetReviewState(int64(orderID), userID, u.Role)
	if err != nil {
		return helper.WriteServiceErrorWithNotFound(c, err, "failed to fetch review state", "order not found")
	}
	return c.JSON(item)
}

func (h *OrderHandler) CreateReview(c *fiber.Ctx) error {
	type reqBody struct {
		Rating    int                `json:"rating"`
		Comment   string             `json:"comment"`
		ImageURLs domain.StringArray `json:"image_urls"`
	}
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	orderID, err := c.ParamsInt("order_id")
	if err != nil || orderID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order_id"})
	}
	var req reqBody
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	item, err := h.service.CreateReview(int64(orderID), userID, u.Role, orderservice.CreateOrderReviewInput{
		Rating:    req.Rating,
		Comment:   req.Comment,
		ImageURLs: req.ImageURLs,
	})
	if err != nil {
		return helper.WriteServiceErrorWithNotFound(c, err, "failed to create review", "order not found",
			helper.BadRequestCase(orderservice.ErrReviewRatingInvalid),
			helper.BadRequestCase(orderservice.ErrReviewCommentInvalid),
			helper.BadRequestCase(orderservice.ErrReviewImagesInvalid),
			helper.ConflictCase(orderservice.ErrReviewOrderNotCompleted),
			helper.ConflictCase(orderservice.ErrReviewAlreadyExists),
		)
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}
