package boq

import (
	"errors"

	"github.com/yourusername/wemake/internal/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	boqservice "github.com/yourusername/wemake/internal/service/boq"
)

type BOQHandler struct {
	service *boqservice.BOQService
}

func NewBOQHandler(service *boqservice.BOQService) *BOQHandler {
	return &BOQHandler{service: service}
}

type boqPayload struct {
	Items          []domain.RFQItem `json:"items"`
	Currency       string           `json:"currency"`
	DiscountAmount float64          `json:"discount_amount"`
	VatPercent     float64          `json:"vat_percent"`
	MOQ            *int             `json:"moq"`
	LeadTimeDays   *int             `json:"lead_time_days"`
	PaymentTerms   *string          `json:"payment_terms"`
	ValidityDays   *int             `json:"validity_days"`
	Note           *string          `json:"note"`
}

func (h *BOQHandler) Create(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	convID, err := helper.ParsePositiveInt64Param(c, "conv_id")
	if err != nil || convID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid conv_id"})
	}
	var req boqPayload
	if err := helper.ParseBody(c, &req, "invalid payload"); err != nil {
		return err
	}
	boq, msg, err := h.service.Create(convID, userID, boqservice.BOQInput(req))
	if err != nil {
		return mapBOQError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"boq": boq, "message": msg})
}

func (h *BOQHandler) Get(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	rfqID, err := helper.ParsePositiveInt64Param(c, "rfq_id")
	if err != nil || rfqID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
	}
	boq, _, err := h.service.Get(rfqID, userID)
	if err != nil {
		return mapBOQError(c, err)
	}
	return c.JSON(boq)
}

func (h *BOQHandler) Update(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	rfqID, err := helper.ParsePositiveInt64Param(c, "rfq_id")
	if err != nil || rfqID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
	}
	var req boqPayload
	if err := helper.ParseBody(c, &req, "invalid payload"); err != nil {
		return err
	}
	boq, err := h.service.Update(rfqID, userID, boqservice.BOQInput(req))
	if err != nil {
		return mapBOQError(c, err)
	}
	return c.JSON(boq)
}

func (h *BOQHandler) Accept(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	rfqID, err := helper.ParsePositiveInt64Param(c, "rfq_id")
	if err != nil || rfqID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
	}
	order, quotationID, err := h.service.Accept(rfqID, userID)
	if err != nil {
		return mapBOQError(c, err)
	}
	return c.JSON(fiber.Map{
		"order_id":     order.OrderID,
		"quotation_id": quotationID,
		"boq_rfq_id":   rfqID,
		"total_amount": order.TotalAmount,
		"status":       order.Status,
		"message":      "BOQ ยืนยันแล้ว กรุณาชำระเงิน",
	})
}

func (h *BOQHandler) Decline(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	rfqID, err := helper.ParsePositiveInt64Param(c, "rfq_id")
	if err != nil || rfqID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
	}
	var req struct {
		Reason *string `json:"reason"`
	}
	_ = helper.ParseBody(c, &req, "invalid payload")
	rfq, err := h.service.Decline(rfqID, userID, req.Reason)
	if err != nil {
		return mapBOQError(c, err)
	}
	return c.JSON(fiber.Map{
		"rfq_id":           rfq.RFQID,
		"boq_response":     rfq.BOQResponse,
		"boq_responded_at": rfq.BOQRespondedAt,
	})
}

func (h *BOQHandler) ListMine(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	items, err := h.service.ListMine(userID, c.Query("status"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch boqs"})
	}
	return c.JSON(items)
}

func mapBOQError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, boqservice.ErrBOQNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "BOQ_NOT_FOUND"})
	case errors.Is(err, boqservice.ErrBOQForbidden):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "BOQ_FORBIDDEN"})
	case errors.Is(err, boqservice.ErrBOQInvalidItems):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "BOQ_INVALID_ITEMS"})
	case errors.Is(err, boqservice.ErrBOQExpired):
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "BOQ_EXPIRED"})
	case errors.Is(err, boqservice.ErrBOQAlreadyHandled):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "BOQ_ALREADY_HANDLED"})
	case errors.Is(err, boqservice.ErrBOQInvalidState):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "BOQ_INVALID_STATE"})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to process boq"})
	}
}
