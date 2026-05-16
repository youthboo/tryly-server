package quotation

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	"github.com/yourusername/wemake/internal/service"
	quotationservice "github.com/yourusername/wemake/internal/service/quotation"
)

type QuotationHandler struct {
	service *quotationservice.QuotationService
	auth    *service.AuthService
}

func NewQuotationHandler(quotationService *quotationservice.QuotationService, authService *service.AuthService) *QuotationHandler {
	return &QuotationHandler{service: quotationService, auth: authService}
}

func (h *QuotationHandler) CreateQuotation(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}

	rfqID, err := c.ParamsInt("rfq_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
	}

	var req dto.CreateQuotationRequest
	if err := helper.ParseAndValidateBody(c, &req, map[string]string{
		"FactoryID":     "factory_id and lead_time_days are required; price_per_piece must be >= 0",
		"PricePerPiece": "factory_id and lead_time_days are required; price_per_piece must be >= 0",
		"LeadTimeDays":  "factory_id and lead_time_days are required; price_per_piece must be >= 0",
	}); err != nil {
		return err
	}
	if req.FactoryID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "factory_id must match authenticated user"})
	}

	// tooling_mold_cost takes precedence over legacy mold_cost
	moldCost := req.MoldCost
	if req.ToolingMoldCost > 0 {
		moldCost = req.ToolingMoldCost
	}
	validityDays := req.ValidityDays
	if validityDays <= 0 {
		validityDays = 14
	}

	item := &domain.Quotation{
		RFQID:            int64(rfqID),
		FactoryID:        req.FactoryID,
		PricePerPiece:    helper.MoneyDecimal(req.PricePerPiece),
		MoldCost:         helper.MoneyDecimal(moldCost),
		ToolingMoldCost:  helper.MoneyDecimal(req.ToolingMoldCost),
		ShippingCost:     helper.MoneyDecimal(req.ShippingCost),
		PackagingCost:    helper.MoneyDecimal(req.PackagingCost),
		LeadTimeDays:     req.LeadTimeDays,
		ValidityDays:     validityDays,
		ShippingMethodID: req.ShippingMethodID,
		PaymentTerms:     req.PaymentTerms,
		ImageURLs:        req.ImageURLs,
		FactoryHighlight: req.FactoryHighlight,
	}
	if err := h.service.Create(item); err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to create quotation"), createQuotationErrorMap())
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *QuotationHandler) ListQuotationsByRFQ(c *fiber.Ctx) error {
	rfqID, err := c.ParamsInt("rfq_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
	}
	items, err := h.service.ListByRFQID(int64(rfqID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch quotations"})
	}
	rfqMeta := fiber.Map{"rfq_id": int64(rfqID)}
	if len(items) > 0 {
		rfqMeta["request_kind"] = items[0].RequestKind
		rfqMeta["sample_qty"] = items[0].SampleQty
		rfqMeta["status"] = items[0].RFQStatus
	}
	return c.JSON(fiber.Map{
		"rfq":        rfqMeta,
		"quotations": items,
	})
}

func (h *QuotationHandler) ListMine(c *fiber.Ctx) error {
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
	status := c.Query("status")
	items, err := h.service.ListMine(userID, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch quotations"})
	}
	return c.JSON(items)
}

func (h *QuotationHandler) ListCollection(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	factoryParam := strings.TrimSpace(c.Query("factory_id"))
	if factoryParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "factory_id query is required"})
	}
	if u.Role != domain.RoleFactory {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "factory role required"})
	}
	var factoryID int64
	if strings.EqualFold(factoryParam, "me") {
		factoryID = userID
	} else {
		parsed, parseErr := helper.ParsePositiveInt64Value(factoryParam, "factory_id")
		if parseErr != nil || parsed <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid factory_id"})
		}
		if parsed != userID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "factory_id must match authenticated factory"})
		}
		factoryID = parsed
	}
	items, err := h.service.ListMine(factoryID, c.Query("status"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch quotations"})
	}
	return c.JSON(items)
}

func (h *QuotationHandler) GetQuotation(c *fiber.Ctx) error {
	quotationID, err := c.ParamsInt("quotation_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid quotation_id"})
	}
	item, err := h.service.GetByID(int64(quotationID))
	if err != nil {
		if isNotFoundError(err) {
			return helper.WriteAPIError(c, helper.NotFoundAPIError("QUOTATION_NOT_FOUND", "quotation not found"))
		}
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_FAILED", "failed to fetch quotation"))
	}
	return c.JSON(item)
}

func (h *QuotationHandler) ListHistory(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	quotationID, err := c.ParamsInt("quotation_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid quotation_id"})
	}
	ok, err := h.service.CanView(int64(quotationID), userID, u.Role)
	if err != nil {
		if isNotFoundError(err) {
			return helper.WriteAPIError(c, helper.NotFoundAPIError("QUOTATION_NOT_FOUND", "quotation not found"))
		}
		return helper.WriteAPIError(c, helper.InternalServerError("AUTH_FAILED", "failed to authorize"))
	}
	if !ok {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not authorized"})
	}
	items, err := h.service.ListHistory(int64(quotationID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch history"})
	}
	return c.JSON(items)
}

func (h *QuotationHandler) Preview(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	var req dto.PreviewQuotationRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	item, err := h.service.Preview(req.Items, req.DiscountAmount, req.ShippingCost, req.PackagingCost, req.ToolingMoldCost, &userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(item)
}

func (h *QuotationHandler) CreateDetailed(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	var req dto.CreateDetailedQuotationRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	item := &domain.Quotation{
		RFQID:                req.RFQID,
		FactoryID:            userID,
		Items:                req.Items,
		DiscountAmount:       helper.MoneyDecimal(req.DiscountAmount),
		ShippingCost:         helper.MoneyDecimal(req.ShippingCost),
		ShippingMethod:       req.ShippingMethod,
		PackagingCost:        helper.MoneyDecimal(req.PackagingCost),
		ToolingMoldCost:      helper.MoneyDecimal(req.ToolingMoldCost),
		Incoterms:            req.Incoterms,
		PaymentTerms:         req.PaymentTerms,
		ValidityDays:         req.ValidityDays,
		WarrantyPeriodMonths: req.WarrantyPeriodMonths,
		FactoryHighlight:     req.FactoryHighlight,
	}
	if req.LeadTimeDays != nil {
		item.LeadTimeDays = *req.LeadTimeDays
	}
	if req.ProductionStartDate != nil && strings.TrimSpace(*req.ProductionStartDate) != "" {
		d, err := helper.ParseDate(*req.ProductionStartDate, "production_start_date")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "production_start_date must be YYYY-MM-DD"})
		}
		item.ProductionStartDate = &d
	}
	if req.DeliveryDate != nil && strings.TrimSpace(*req.DeliveryDate) != "" {
		d, err := helper.ParseDate(*req.DeliveryDate, "delivery_date")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "delivery_date must be YYYY-MM-DD"})
		}
		item.DeliveryDate = &d
	}
	if err := h.service.CreateDetailed(item); err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusBadRequest, "failed to create detailed quotation"), createDetailedErrorMap())
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *QuotationHandler) CreateRevision(c *fiber.Ctx) error {
	parentID, err := c.ParamsInt("quotation_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid quotation_id"})
	}
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	var req dto.CreateRevisionQuotationRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	item := &domain.Quotation{
		FactoryID:            userID,
		Items:                req.Items,
		DiscountAmount:       helper.MoneyDecimal(req.DiscountAmount),
		ShippingCost:         helper.MoneyDecimal(req.ShippingCost),
		ShippingMethod:       req.ShippingMethod,
		PackagingCost:        helper.MoneyDecimal(req.PackagingCost),
		ToolingMoldCost:      helper.MoneyDecimal(req.ToolingMoldCost),
		Incoterms:            req.Incoterms,
		PaymentTerms:         req.PaymentTerms,
		ValidityDays:         req.ValidityDays,
		WarrantyPeriodMonths: req.WarrantyPeriodMonths,
	}
	if req.LeadTimeDays != nil {
		item.LeadTimeDays = *req.LeadTimeDays
	}
	if err := h.service.CreateRevision(int64(parentID), userID, item); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *QuotationHandler) Accept(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	quotationID, err := c.ParamsInt("quotation_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid quotation_id"})
	}
	order, err := h.service.Accept(int64(quotationID), userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(order)
}

func (h *QuotationHandler) Reject(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	quotationID, err := c.ParamsInt("quotation_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid quotation_id"})
	}
	if err := h.service.Reject(int64(quotationID), userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "quotation rejected"})
}

func (h *QuotationHandler) PatchQuotation(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	quotationID, err := c.ParamsInt("quotation_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid quotation_id"})
	}
	var req dto.PatchQuotationBodyRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	// tooling_mold_cost takes precedence over legacy mold_cost
	moldCost := req.MoldCost
	if req.ToolingMoldCost > 0 {
		moldCost = req.ToolingMoldCost
	}
	item, err := h.service.PatchBody(
		int64(quotationID),
		userID,
		req.PricePerPiece,
		moldCost,
		req.ShippingCost,
		req.PackagingCost,
		req.ToolingMoldCost,
		req.LeadTimeDays,
		req.ShippingMethodID,
		req.PaymentTerms,
		req.FactoryHighlight,
		req.Reason,
		req.ValidityDays,
	)
	if err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to update quotation"), patchQuotationErrorMap())
	}
	// Update image_urls if provided in request
	if req.ImageURLs != nil {
		if imgErr := h.service.UpdateImageURLs(int64(quotationID), req.ImageURLs); imgErr != nil {
			// non-fatal — log but return existing item
			_ = imgErr
		} else {
			item.ImageURLs = req.ImageURLs
		}
	}
	return c.JSON(item)
}

func (h *QuotationHandler) PatchQuotationStatus(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	var editor *int64
	if err == nil {
		editor = &userID
	}
	quotationID, err := c.ParamsInt("quotation_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid quotation_id"})
	}
	var req dto.PatchQuotationStatusRequest
	if err := helper.ParseAndValidateBody(c, &req, map[string]string{
		"Status": "status must be PD, AC, RJ or EX",
	}); err != nil {
		return err
	}
	status := strings.TrimSpace(strings.ToUpper(req.Status))
	if status != "AC" && status != "RJ" && status != "PD" && status != "EX" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be PD, AC, RJ or EX"})
	}
	if err := h.service.UpdateStatus(int64(quotationID), status, editor); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update quotation status"})
	}
	return c.JSON(fiber.Map{"message": "quotation status updated"})
}
