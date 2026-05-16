package payment

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/dto"
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
	orderID, err := helper.RequireInt64Param(c, "order_id")
	if err != nil {
		return err
	}
	items, err := h.service.ListByOrderID(int64(orderID))
	if err != nil {
		return helper.InternalServerError(c, "failed to fetch payment schedules")
	}
	return c.JSON(items)
}

// POST /orders/:order_id/payment-schedules
func (h *PaymentScheduleHandler) Create(c *fiber.Ctx) error {
	orderID, err := helper.RequireInt64Param(c, "order_id")
	if err != nil {
		return err
	}
	var req dto.CreatePaymentScheduleRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	v := domain.NewValidationCollector()
	v.AddIf(req.InstallmentNo <= 0, "installment_no", "is required")
	v.AddIf(helper.DereferenceString(&req.DueDate, "") == "", "due_date", "is required")
	v.AddIf(req.Amount <= 0, "amount", "must be positive")
	if err := helper.ValidateRequest(c, v); err != nil {
		return err
	}
	dueDate, err := helper.ParseRequiredDateValue(req.DueDate, "due_date")
	if err != nil {
		return helper.BadRequest(c, "due_date must be YYYY-MM-DD")
	}
	item := &domain.PaymentSchedule{
		OrderID:       int64(orderID),
		InstallmentNo: req.InstallmentNo,
		DueDate:       dueDate,
		Amount:        helper.MoneyDecimal(req.Amount),
	}
	if err := h.service.CreateSchedule(item); err != nil {
		return helper.InternalServerError(c, "failed to create payment schedule")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

// PATCH /payment-schedules/:schedule_id
func (h *PaymentScheduleHandler) PatchStatus(c *fiber.Ctx) error {
	scheduleID, err := helper.ParsePositiveInt64Param(c, "schedule_id")
	if err != nil {
		return helper.BadRequest(c, "invalid schedule_id")
	}
	var req dto.PatchPaymentScheduleStatusRequest
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
