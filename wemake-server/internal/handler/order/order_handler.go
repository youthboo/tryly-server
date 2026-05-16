package order

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/dbutil"
	"github.com/yourusername/wemake/internal/domain"
	domainstatus "github.com/yourusername/wemake/internal/domain/status"
	"github.com/yourusername/wemake/internal/dto"
	handlerregistry "github.com/yourusername/wemake/internal/handler/errorregistry"
	"github.com/yourusername/wemake/internal/helper"
	authservice "github.com/yourusername/wemake/internal/service/auth"
	orderservice "github.com/yourusername/wemake/internal/service/order"
)

type OrderHandler struct {
	service *orderservice.OrderService
	auth    *authservice.AuthService
}

func NewOrderHandler(orderService *orderservice.OrderService, authService *authservice.AuthService) *OrderHandler {
	return &OrderHandler{service: orderService, auth: authService}
}

func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	userID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}
	var req dto.CreateOrderFromQuoteRequest
	if err := helper.ParseAndValidateBody(c, &req, map[string]string{
		"QuotationID": "quote_id is required",
	}); err != nil {
		return err
	}
	order, err := h.service.CreateFromQuotation(req.QuotationID, userID)
	if err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to create order"), handlerregistry.CreateOrderErrorMap())
	}
	return c.Status(fiber.StatusCreated).JSON(order)
}

func (h *OrderHandler) BulkCheckout(c *fiber.Ctx) error {
	userID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}
	rfqID, err := helper.RequirePathID(c, "rfq_id")
	if err != nil {
		return err
	}
	var req dto.BulkCheckoutBodyRequest
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
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to bulk checkout"), handlerregistry.BulkCheckoutErrorMap())
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *OrderHandler) ListOrders(c *fiber.Ctx) error {
	userID, u, err := helper.RequireUser(c, h.auth)
	if err != nil {
		return err
	}
	if helper.QueryString(c, "factory_id") != "" {
		if err := helper.RequireFactoryRole(u); err != nil {
			return helper.JSONError(c, fiber.StatusForbidden, "factory role required")
		}
		if _, err := helper.QueryMatchingSelfOrID(c, "factory_id", userID, "factory_id must match authenticated factory"); err != nil {
			return err
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
		return helper.JSONInternal(c, "failed to fetch orders")
	}
	return c.JSON(items)
}

func (h *OrderHandler) GetOrder(c *fiber.Ctx) error {
	userID, u, err := helper.RequireUser(c, h.auth)
	if err != nil {
		return err
	}
	orderID, err := helper.RequirePathID(c, "order_id")
	if err != nil {
		return err
	}
	detail, err := h.service.GetDetailByID(int64(orderID), userID, u.Role)
	if err != nil {
		return helper.WriteServiceErrorWithNotFound(c, err, "failed to fetch order", "order not found")
	}
	return c.JSON(detail)
}

func (h *OrderHandler) ListActivity(c *fiber.Ctx) error {
	userID, u, err := helper.RequireUser(c, h.auth)
	if err != nil {
		return err
	}
	orderID, err := helper.RequirePathID(c, "order_id")
	if err != nil {
		return err
	}
	if _, err := h.service.GetByID(int64(orderID), userID, u.Role); err != nil {
		return helper.WriteServiceErrorWithNotFound(c, err, "failed to verify order", "order not found")
	}
	items, err := h.service.ListActivity(int64(orderID))
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch activity")
	}
	return c.JSON(items)
}

func (h *OrderHandler) PatchOrderStatus(c *fiber.Ctx) error {
	uid, authErr := helper.UserIDFromHeader(c)
	var actor *int64
	if authErr == nil {
		actor = &uid
	}
	orderID, err := helper.RequirePathID(c, "order_id")
	if err != nil {
		return err
	}
	var req dto.PatchOrderStatusRequest
	if err := helper.ParseAndValidateBody(c, &req, map[string]string{
		"Status": "status must be PP, PR, WF, QC, SH, DL, AC, CP, or CC",
	}); err != nil {
		return err
	}
	status := domainstatus.NormalizeOrder(req.Status)
	if !domainstatus.IsValidOrder(status) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be PP, PR, WF, QC, SH, DL, AC, CP, or CC"})
	}
	if err := h.service.UpdateStatus(int64(orderID), status, actor); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update order status"})
	}
	return c.JSON(fiber.Map{"message": "order status updated"})
}

func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	userID, u, err := helper.RequireUser(c, h.auth)
	if err != nil {
		return err
	}
	orderID, err := helper.RequirePathID(c, "order_id")
	if err != nil {
		return err
	}
	if err := h.service.Cancel(int64(orderID), userID, u.Role); err != nil {
		return helper.WriteServiceErrorWithNotFound(c, err, "failed to cancel order", "order not found", helper.BadRequestCase(orderservice.ErrOrderCannotBeCancelled))
	}
	return c.JSON(fiber.Map{"message": "order cancelled"})
}

func (h *OrderHandler) MarkShipped(c *fiber.Ctx) error {
	userID, u, err := helper.RequireFactoryUser(c, h.auth)
	if err != nil {
		return err
	}
	orderID, err := helper.RequirePathID(c, "order_id")
	if err != nil {
		return err
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
	userID, u, err := helper.RequireUser(c, h.auth)
	if err != nil {
		return err
	}
	orderID, err := helper.RequirePathID(c, "order_id")
	if err != nil {
		return err
	}
	var req dto.PayDepositRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	item, err := h.service.CreatePayment(int64(orderID), userID, u.Role, req.Type, req.Amount)
	if err != nil {
		errorMap := handlerregistry.CreatePaymentErrorMap()
		if dbutil.IsNotFoundError(err) {
			return helper.WriteAPIError(c, helper.NotFoundAPIError("ORDER_NOT_FOUND", "order not found"))
		}
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to create payment"), errorMap)
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *OrderHandler) VerifyPayment(c *fiber.Ctx) error {
	userID, u, err := helper.RequireUser(c, h.auth)
	if err != nil {
		return err
	}
	orderID, err := helper.RequirePathID(c, "order_id")
	if err != nil {
		return err
	}
	txID := strings.TrimSpace(c.Params("tx_id"))
	if txID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tx_id"})
	}
	item, err := h.service.VerifyPayment(int64(orderID), userID, u.Role, txID)
	if err != nil {
		errorMap := handlerregistry.VerifyPaymentErrorMap()
		if dbutil.IsNotFoundError(err) {
			return helper.WriteAPIError(c, helper.NotFoundAPIError("PAYMENT_NOT_FOUND", "payment not found"))
		}
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to verify payment"), errorMap)
	}
	return c.JSON(item)
}

func (h *OrderHandler) ConfirmReceipt(c *fiber.Ctx) error {
	userID, u, err := helper.RequireUser(c, h.auth)
	if err != nil {
		return err
	}
	orderID, err := helper.RequirePathID(c, "order_id")
	if err != nil {
		return err
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
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to confirm receipt"), handlerregistry.ConfirmReceiptErrorMap())
	}
	return c.JSON(result)
}

func (h *OrderHandler) GetReviewState(c *fiber.Ctx) error {
	userID, u, err := helper.RequireUser(c, h.auth)
	if err != nil {
		return err
	}
	orderID, err := helper.RequirePathID(c, "order_id")
	if err != nil {
		return err
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
	userID, u, err := helper.RequireUser(c, h.auth)
	if err != nil {
		return err
	}
	orderID, err := helper.RequirePathID(c, "order_id")
	if err != nil {
		return err
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
