package boq

import (
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"

	"github.com/gofiber/fiber/v2"
	boqservice "github.com/yourusername/wemake/internal/service/boq"
)

type BOQHandler struct {
	service *boqservice.BOQService
}

func NewBOQHandler(service *boqservice.BOQService) *BOQHandler {
	return &BOQHandler{service: service}
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
	var req dto.BOQPayloadRequest
	if err := helper.ParseBody(c, &req, "invalid payload"); err != nil {
		return err
	}
	input := boqPayloadToInput(req)
	boq, msg, err := h.service.Create(convID, userID, input)
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
	var req dto.BOQPayloadRequest
	if err := helper.ParseBody(c, &req, "invalid payload"); err != nil {
		return err
	}
	input := boqPayloadToInput(req)
	boq, err := h.service.Update(rfqID, userID, input)
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
	var req dto.DeclineBOQRequest
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

func boqPayloadToInput(req dto.BOQPayloadRequest) boqservice.BOQInput {
	input := boqservice.BOQInput{
		Currency:       req.Currency,
		DiscountAmount: req.DiscountAmount,
		VatPercent:     req.VatPercent,
		MOQ:            req.MOQ,
		LeadTimeDays:   req.LeadTimeDays,
		PaymentTerms:   req.PaymentTerms,
		ValidityDays:   req.ValidityDays,
		Note:           req.Note,
	}
	// Convert interface{} items to domain.RFQItem
	for _, item := range req.Items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		rfqItem := domain.RFQItem{}
		if desc, ok := itemMap["description"].(string); ok {
			rfqItem.Description = desc
		}
		if qty, ok := itemMap["qty"].(float64); ok {
			rfqItem.Qty = qty
		}
		if unitPrice, ok := itemMap["unit_price"].(float64); ok {
			rfqItem.UnitPrice = unitPrice
		}
		if discountPct, ok := itemMap["discount_pct"].(float64); ok {
			rfqItem.DiscountPct = discountPct
		}
		if unit, ok := itemMap["unit"].(string); ok {
			rfqItem.Unit = &unit
		}
		if spec, ok := itemMap["specification"].(string); ok {
			rfqItem.Specification = &spec
		}
		input.Items = append(input.Items, rfqItem)
	}
	return input
}

func mapBOQError(c *fiber.Ctx, err error) error {
	return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to process boq"), boqErrorMap())
}
