package rfq

import (
	"database/sql"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	authservice "github.com/yourusername/wemake/internal/service/auth"
	rfqservice "github.com/yourusername/wemake/internal/service/rfq"
)

type RFQHandler struct {
	service *rfqservice.RFQService
	auth    *authservice.AuthService
}

func NewRFQHandler(rfqService *rfqservice.RFQService, authService *authservice.AuthService) *RFQHandler {
	return &RFQHandler{service: rfqService, auth: authService}
}

var rfqCreateErrorMap = map[error]helper.ErrorResponse{
	rfqservice.ErrInvalidSubCategory:    helper.ErrorMessage(fiber.StatusBadRequest, rfqservice.ErrInvalidSubCategory.Error()),
	rfqservice.ErrInvalidCategory:       helper.ErrorMessage(fiber.StatusBadRequest, rfqservice.ErrInvalidCategory.Error()),
	rfqservice.ErrInvalidShippingMethod: helper.ErrorMessage(fiber.StatusBadRequest, rfqservice.ErrInvalidShippingMethod.Error()),
	rfqservice.ErrMaxRFQReferenceImages: helper.ErrorMessage(fiber.StatusBadRequest, rfqservice.ErrMaxRFQReferenceImages.Error()),
	rfqservice.ErrRFQInspectionInvalid:  helper.ErrorMessage(fiber.StatusBadRequest, rfqservice.ErrRFQInspectionInvalid.Error()),
	rfqservice.ErrRFQDetailsRequired:    helper.ErrorMessage(fiber.StatusBadRequest, rfqservice.ErrRFQDetailsRequired.Error()),
	rfqservice.ErrRFQDetailsTooShort:    helper.ErrorMessage(fiber.StatusBadRequest, rfqservice.ErrRFQDetailsTooShort.Error()),
	rfqservice.ErrRFQKindInvalid:        helper.ErrorMessage(fiber.StatusBadRequest, rfqservice.ErrRFQKindInvalid.Error()),
	rfqservice.ErrRFQSampleQtyInvalid:   helper.ErrorMessage(fiber.StatusBadRequest, rfqservice.ErrRFQSampleQtyInvalid.Error()),
	rfqservice.ErrRFQWrongScope:         helper.ErrorMessage(fiber.StatusBadRequest, rfqservice.ErrRFQWrongScope.Error()),
}

var rfqPreviewErrorMap = map[error]helper.ErrorResponse{
	rfqservice.ErrRFQKindInvalid:     helper.ErrorMessage(fiber.StatusBadRequest, "INVALID_KIND"),
	rfqservice.ErrRFQWrongScope:      helper.ErrorMessage(fiber.StatusBadRequest, "WRONG_SCOPE"),
	rfqservice.ErrInvalidSubCategory: helper.ErrorMessage(fiber.StatusNotFound, "CATEGORY_NOT_FOUND"),
	rfqservice.ErrInvalidCategory:    helper.ErrorMessage(fiber.StatusNotFound, "CATEGORY_NOT_FOUND"),
}

var rfqDismissErrorMap = map[error]helper.ErrorResponse{
	rfqservice.ErrHasActiveQuotation: helper.ErrorMessage(fiber.StatusConflict, "HAS_ACTIVE_QUOTATION"),
	rfqservice.ErrQuotationAccepted:  helper.ErrorMessage(fiber.StatusConflict, "QUOTATION_ACCEPTED"),
	sql.ErrNoRows:                    helper.ErrorMessage(fiber.StatusNotFound, "RFQ_NOT_FOUND"),
}

var rfqNotFoundErrorMap = map[error]helper.ErrorResponse{
	sql.ErrNoRows: helper.ErrorMessage(fiber.StatusNotFound, "rfq not found"),
}

var rfqNotFoundCodeErrorMap = map[error]helper.ErrorResponse{
	sql.ErrNoRows: helper.ErrorMessage(fiber.StatusNotFound, "RFQ_NOT_FOUND"),
}

func (h *RFQHandler) CreateRFQ(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}

	var req dto.CreateRFQRequest
	if err := helper.ParseAndValidateBody(c, &req, map[string]string{
		"CategoryID": "category_id, title, quantity, and address_id are required",
		"Title":      "category_id, title, quantity, and address_id are required",
		"Quantity":   "category_id, title, quantity, and address_id are required",
		"AddressID":  "category_id, title, quantity, and address_id are required",
	}); err != nil {
		return err
	}

	details := strings.TrimSpace(req.Details)
	if details == "" {
		details = strings.TrimSpace(req.Description)
	}

	rfq := &domain.RFQ{
		UserID:                 userID,
		CategoryID:             req.CategoryID,
		SubCategoryID:          req.SubCategoryID,
		Title:                  req.Title,
		Quantity:               req.Quantity,
		Details:                details,
		AddressID:              req.AddressID,
		ShippingMethodID:       req.ShippingMethodID,
		MaterialGrade:          req.MaterialGrade,
		TargetPrice:            helper.MoneyDecimalPtr(req.TargetPrice),
		TargetLeadTimeDays:     req.TargetLeadTimeDays,
		DeliveryAddressID:      req.DeliveryAddressID,
		CertificationsRequired: req.CertificationsRequired,
		SampleRequired:         req.SampleRequired,
		SampleQty:              req.SampleQty,
		InspectionType:         req.InspectionType,
		ReferenceImages:        req.ReferenceImages,
		RequestKind:            req.RequestKind,
	}
	if rfq.DeliveryAddressID == nil {
		rfq.DeliveryAddressID = &rfq.AddressID
	}
	if req.RequiredDeliveryDate != nil && strings.TrimSpace(*req.RequiredDeliveryDate) != "" {
		d, err := helper.ParseDate(*req.RequiredDeliveryDate, "required_delivery_date")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "required_delivery_date must be YYYY-MM-DD"})
		}
		rfq.RequiredDeliveryDate = &d
	}

	if err := h.service.Create(rfq); err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to create rfq"), rfqCreateErrorMap)
	}
	domain.EnrichRFQBudgetFields(rfq)
	return c.Status(fiber.StatusCreated).JSON(rfq)
}

func (h *RFQHandler) PatchRFQ(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	rfqID, err := c.ParamsInt("rfq_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
	}
	var req dto.PatchRFQRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	details := ""
	if req.Details != nil {
		details = strings.TrimSpace(*req.Details)
	}
	if details == "" && req.Description != nil {
		details = strings.TrimSpace(*req.Description)
	}
	rfq := &domain.RFQ{
		Details: details,
	}
	if req.CategoryID != nil {
		rfq.CategoryID = *req.CategoryID
	}
	if req.SubCategoryID != nil {
		rfq.SubCategoryID = req.SubCategoryID
	}
	if req.Title != nil {
		rfq.Title = *req.Title
	}
	if req.Quantity != nil {
		rfq.Quantity = *req.Quantity
	}
	if req.MaterialGrade != nil {
		rfq.MaterialGrade = req.MaterialGrade
	}
	if req.TargetPrice != nil {
		rfq.TargetPrice = helper.MoneyDecimalPtr(req.TargetPrice)
	}
	if req.TargetLeadTimeDays != nil {
		rfq.TargetLeadTimeDays = req.TargetLeadTimeDays
	}
	if req.SampleRequired != nil {
		rfq.SampleRequired = *req.SampleRequired
	}
	if req.SampleQty != nil {
		rfq.SampleQty = req.SampleQty
	}
	if req.InspectionType != nil {
		rfq.InspectionType = req.InspectionType
	}
	if req.ReferenceImages != nil {
		rfq.ReferenceImages = req.ReferenceImages
	}
	if req.RequiredDeliveryDate != nil && strings.TrimSpace(*req.RequiredDeliveryDate) != "" {
		d, err := helper.ParseDate(*req.RequiredDeliveryDate, "required_delivery_date")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "required_delivery_date must be YYYY-MM-DD"})
		}
		rfq.RequiredDeliveryDate = &d
	}
	if err := h.service.Patch(userID, int64(rfqID), rfq); err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusBadRequest, err.Error()), rfqCreateErrorMap)
	}
	domain.EnrichRFQBudgetFields(rfq)
	return c.JSON(rfq)
}

func (h *RFQHandler) ListRFQs(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	status := c.Query("status")
	rfqs, err := h.service.ListByUserID(userID, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch rfqs"})
	}
	kind := strings.TrimSpace(strings.ToUpper(c.Query("kind")))
	if kind != "" {
		if kind != domain.RequestKindProduction && kind != domain.RequestKindProductSample && kind != domain.RequestKindMaterialSample {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "INVALID_KIND"})
		}
		filtered := make([]domain.RFQ, 0, len(rfqs))
		for _, item := range rfqs {
			if strings.EqualFold(item.RequestKind, kind) {
				filtered = append(filtered, item)
			}
		}
		rfqs = filtered
	}
	return c.JSON(rfqs)
}

func (h *RFQHandler) PreviewFactories(c *fiber.Ctx) error {
	kind := strings.TrimSpace(c.Query("kind"))
	if kind == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "INVALID_KIND"})
	}
	rawCategory := strings.TrimSpace(c.Query("category_id"))
	if rawCategory == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "MISSING_CATEGORY"})
	}
	categoryID, err := helper.ParsePositiveInt64Value(rawCategory, "category_id")
	if err != nil || categoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "MISSING_CATEGORY"})
	}
	var subCategoryID *int64
	if raw := strings.TrimSpace(c.Query("sub_category_id")); raw != "" {
		parsed, parseErr := helper.ParsePositiveInt64Value(raw, "sub_category_id")
		if parseErr != nil || parsed <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid sub_category_id"})
		}
		subCategoryID = &parsed
	}
	result, err := h.service.PreviewFactories(kind, categoryID, subCategoryID)
	if err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to preview factories"), rfqPreviewErrorMap)
	}
	return c.JSON(result)
}

func (h *RFQHandler) ListMatching(c *fiber.Ctx) error {
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
	showDismissed := strings.EqualFold(strings.TrimSpace(c.Query("show_dismissed")), "true")
	items, err := h.service.ListMatchingForFactory(userID, status, c.Query("kind"), showDismissed)
	if err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to fetch matching rfqs"), map[error]helper.ErrorResponse{
			rfqservice.ErrRFQKindInvalid: helper.ErrorMessage(fiber.StatusBadRequest, "INVALID_KIND"),
		})
	}
	return c.JSON(items)
}

func (h *RFQHandler) DismissRFQ(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	if u.Role != domain.RoleFactory {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "FORBIDDEN"})
	}
	rfqID, err := c.ParamsInt("rfq_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
	}
	item, created, err := h.service.DismissRFQ(userID, int64(rfqID))
	if err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to dismiss rfq"), rfqDismissErrorMap)
	}
	status := fiber.StatusOK
	if created {
		status = fiber.StatusCreated
	}
	return c.Status(status).JSON(fiber.Map{
		"rfq_id":       item.RFQID,
		"dismissed":    true,
		"dismissed_at": item.DismissedAt,
	})
}

func (h *RFQHandler) UndismissRFQ(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	if u.Role != domain.RoleFactory {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "FORBIDDEN"})
	}
	rfqID, err := c.ParamsInt("rfq_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
	}
	if err := h.service.UndismissRFQ(userID, int64(rfqID)); err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to undismiss rfq"), rfqNotFoundCodeErrorMap)
	}
	return c.JSON(fiber.Map{"rfq_id": int64(rfqID), "dismissed": false})
}

func (h *RFQHandler) GetRFQ(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	rfqID, err := c.ParamsInt("rfq_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
	}

	rfq, err := h.service.GetForViewer(userID, u.Role, int64(rfqID))
	if err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to fetch rfq"), rfqNotFoundErrorMap)
	}

	return c.JSON(fiber.Map{"rfq": rfq})
}

func (h *RFQHandler) CancelRFQ(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	rfqID, err := c.ParamsInt("rfq_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
	}

	if err := h.service.Cancel(userID, int64(rfqID)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to cancel rfq"})
	}
	return c.JSON(fiber.Map{"message": "rfq canceled"})
}

// CloseRFQ lets the customer manually close (stop accepting new quotes) an open RFQ.
// PATCH /rfqs/:rfq_id/close
func (h *RFQHandler) CloseRFQ(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	rfqID, err := c.ParamsInt("rfq_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
	}

	if err := h.service.Close(userID, int64(rfqID)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to close rfq"})
	}
	return c.JSON(fiber.Map{"message": "rfq closed"})
}
