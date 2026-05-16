package payment

import (
	"database/sql"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/helper"
	paymentservice "github.com/yourusername/wemake/internal/service/payment"
)

type PaymentScheduleHandler struct {
	service *paymentservice.PaymentScheduleService
}

func NewPaymentScheduleHandler(svc *paymentservice.PaymentScheduleService) *PaymentScheduleHandler {
	return &PaymentScheduleHandler{service: svc}
}

// GET /orders/:order_id/payment-schedules
func (h *PaymentScheduleHandler) List(c *fiber.Ctx) error {
	orderID, err := c.ParamsInt("order_id")
	if err != nil || orderID <= 0 {
		return helper.BadRequest(c, "invalid order_id")
	}
	items, err := h.service.ListByOrderID(int64(orderID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch payment schedules"})
	}
	return c.JSON(items)
}

// POST /orders/:order_id/payment-schedules
func (h *PaymentScheduleHandler) Create(c *fiber.Ctx) error {
	type reqBody struct {
		InstallmentNo int     `json:"installment_no"`
		DueDate       string  `json:"due_date"` // YYYY-MM-DD
		Amount        float64 `json:"amount"`
	}
	orderID, err := c.ParamsInt("order_id")
	if err != nil || orderID <= 0 {
		return helper.BadRequest(c, "invalid order_id")
	}
	var req reqBody
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	if req.InstallmentNo <= 0 || strings.TrimSpace(req.DueDate) == "" || req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "installment_no, due_date, and amount are required"})
	}
	dueDate, err := helper.ParseRequiredDateValue(req.DueDate, "due_date")
	if err != nil {
		return helper.BadRequest(c, "due_date must be YYYY-MM-DD")
	}
	item := &domain.PaymentSchedule{
		OrderID:       int64(orderID),
		InstallmentNo: req.InstallmentNo,
		DueDate:       dueDate,
		Amount:        req.Amount,
	}
	if err := h.service.CreateSchedule(item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create payment schedule"})
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

// PATCH /payment-schedules/:schedule_id
func (h *PaymentScheduleHandler) PatchStatus(c *fiber.Ctx) error {
	type reqBody struct {
		Status string `json:"status"`
	}
	scheduleID, err := helper.ParsePositiveInt64Param(c, "schedule_id")
	if err != nil {
		return helper.BadRequest(c, "invalid schedule_id")
	}
	var req reqBody
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	if err := h.service.PatchStatus(scheduleID, req.Status); err != nil {
		return helper.MapServiceError(c, err, paymentSchedulePatchStatusFallback, paymentSchedulePatchStatusResponses)
	}
	return c.JSON(fiber.Map{"message": "payment schedule updated"})
}

var paymentSchedulePatchStatusFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to update payment schedule")

var paymentSchedulePatchStatusResponses = map[error]helper.ErrorResponse{
	paymentservice.ErrInvalidScheduleStatus: helper.ErrorMessage(fiber.StatusBadRequest, paymentservice.ErrInvalidScheduleStatus.Error()),
	sql.ErrNoRows:                           helper.ErrorMessage(fiber.StatusNotFound, "payment schedule not found"),
}
