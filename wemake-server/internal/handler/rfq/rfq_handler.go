package rfq

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/repository"
	"github.com/yourusername/wemake/internal/service"
	rfqservice "github.com/yourusername/wemake/internal/service/rfq"
)

type RFQHandler struct {
	service *rfqservice.RFQService
	auth    *service.AuthService
}

func NewRFQHandler(rfqService *rfqservice.RFQService, authService *service.AuthService) *RFQHandler {
	return &RFQHandler{service: rfqService, auth: authService}
}

func (h *RFQHandler) CreateRFQ(c *fiber.Ctx) error {
	type createRFQRequest struct {
		CategoryID             int64    `json:"category_id" validate:"gt=0"`
		SubCategoryID          *int64   `json:"sub_category_id"`
		Title                  string   `json:"title" validate:"notblank"`
		Description            string   `json:"description"`
		Quantity               int64    `json:"quantity" validate:"gt=0"`
		Unit                   string   `json:"unit"`
		Details                string   `json:"details"`
		AddressID              int64    `json:"address_id" validate:"gt=0"`
		ShippingMethodID       *int64   `json:"shipping_method_id"`
		MaterialGrade          *string  `json:"material_grade"`
		TargetPrice            *float64 `json:"target_price"`
		TargetLeadTimeDays     *int     `json:"target_lead_time_days"`
		RequiredDeliveryDate   *string  `json:"required_delivery_date"`
		DeliveryAddressID      *int64   `json:"delivery_address_id"`
		CertificationsRequired []string `json:"certifications_required"`
		SampleRequired         bool     `json:"sample_required"`
		SampleQty              *int     `json:"sample_qty"`
		InspectionType         *string  `json:"inspection_type"`
		ReferenceImages        []string `json:"reference_images"`
		RequestKind            string   `json:"request_kind"`
	}

	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}

	var req createRFQRequest
	if err := parseAndValidateBody(c, &req, map[string]string{
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
		TargetPrice:            req.TargetPrice,
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
		d, err := time.Parse("2006-01-02", strings.TrimSpace(*req.RequiredDeliveryDate))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "required_delivery_date must be YYYY-MM-DD"})
		}
		rfq.RequiredDeliveryDate = &d
	}

	if err := h.service.Create(rfq); err != nil {
		if err == rfqservice.ErrInvalidSubCategory || err == rfqservice.ErrInvalidCategory || err == rfqservice.ErrInvalidShippingMethod || err == rfqservice.ErrMaxRFQReferenceImages || err == rfqservice.ErrRFQInspectionInvalid || err == rfqservice.ErrRFQDetailsRequired || err == rfqservice.ErrRFQDetailsTooShort || err == rfqservice.ErrRFQKindInvalid || err == rfqservice.ErrRFQSampleQtyInvalid || err == rfqservice.ErrRFQWrongScope {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create rfq"})
	}
	domain.EnrichRFQBudgetFields(rfq)
	return c.Status(fiber.StatusCreated).JSON(rfq)
}

func (h *RFQHandler) PatchRFQ(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid X-User-ID header"})
	}
	rfqID, err := c.ParamsInt("rfq_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid rfq_id"})
	}
	type patchRFQRequest struct {
		CategoryID             int64    `json:"category_id"`
		SubCategoryID          *int64   `json:"sub_category_id"`
		Title                  string   `json:"title"`
		Description            string   `json:"description"`
		Quantity               int64    `json:"quantity"`
		Unit                   string   `json:"unit"`
		Details                string   `json:"details"`
		AddressID              int64    `json:"address_id"`
		ShippingMethodID       *int64   `json:"shipping_method_id"`
		MaterialGrade          *string  `json:"material_grade"`
		TargetPrice            *float64 `json:"target_price"`
		TargetLeadTimeDays     *int     `json:"target_lead_time_days"`
		RequiredDeliveryDate   *string  `json:"required_delivery_date"`
		DeliveryAddressID      *int64   `json:"delivery_address_id"`
		CertificationsRequired []string `json:"certifications_required"`
		SampleRequired         bool     `json:"sample_required"`
		SampleQty              *int     `json:"sample_qty"`
		InspectionType         *string  `json:"inspection_type"`
		ReferenceImages        []string `json:"reference_images"`
		RequestKind            string   `json:"request_kind"`
	}
	var req patchRFQRequest
	if err := requireBody(c, &req); err != nil {
		return err
	}
	details := strings.TrimSpace(req.Details)
	if details == "" {
		details = strings.TrimSpace(req.Description)
	}
	rfq := &domain.RFQ{
		CategoryID: req.CategoryID, SubCategoryID: req.SubCategoryID, Title: req.Title, Quantity: req.Quantity,
		Details: details, AddressID: req.AddressID,
		ShippingMethodID: req.ShippingMethodID,
		MaterialGrade:    req.MaterialGrade, TargetPrice: req.TargetPrice,
		TargetLeadTimeDays: req.TargetLeadTimeDays,
		DeliveryAddressID:  req.DeliveryAddressID, CertificationsRequired: req.CertificationsRequired, SampleRequired: req.SampleRequired,
		SampleQty: req.SampleQty, InspectionType: req.InspectionType, ReferenceImages: req.ReferenceImages,
		RequestKind: req.RequestKind,
	}
	if req.RequiredDeliveryDate != nil && strings.TrimSpace(*req.RequiredDeliveryDate) != "" {
		d, err := time.Parse("2006-01-02", strings.TrimSpace(*req.RequiredDeliveryDate))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "required_delivery_date must be YYYY-MM-DD"})
		}
		rfq.RequiredDeliveryDate = &d
	}
	if err := h.service.Patch(userID, int64(rfqID), rfq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	domain.EnrichRFQBudgetFields(rfq)
	return c.JSON(rfq)
}

func (h *RFQHandler) ListRFQs(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
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
	categoryID, err := strconv.ParseInt(rawCategory, 10, 64)
	if err != nil || categoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "MISSING_CATEGORY"})
	}
	var subCategoryID *int64
	if raw := strings.TrimSpace(c.Query("sub_category_id")); raw != "" {
		parsed, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || parsed <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid sub_category_id"})
		}
		subCategoryID = &parsed
	}
	result, err := h.service.PreviewFactories(kind, categoryID, subCategoryID)
	if err != nil {
		if err == rfqservice.ErrRFQKindInvalid {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "INVALID_KIND"})
		}
		if err == rfqservice.ErrRFQWrongScope {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "WRONG_SCOPE"})
		}
		if err == rfqservice.ErrInvalidSubCategory || err == rfqservice.ErrInvalidCategory {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "CATEGORY_NOT_FOUND"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to preview factories"})
	}
	return c.JSON(result)
}

func (h *RFQHandler) ListMatching(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
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
		if err == rfqservice.ErrRFQKindInvalid {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "INVALID_KIND"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch matching rfqs"})
	}
	return c.JSON(items)
}

func (h *RFQHandler) DismissRFQ(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
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
		switch err {
		case rfqservice.ErrHasActiveQuotation:
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "HAS_ACTIVE_QUOTATION"})
		case rfqservice.ErrQuotationAccepted:
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "QUOTATION_ACCEPTED"})
		case sql.ErrNoRows:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "RFQ_NOT_FOUND"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to dismiss rfq"})
		}
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
	userID, err := getUserIDFromHeader(c)
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
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "RFQ_NOT_FOUND"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to undismiss rfq"})
	}
	return c.JSON(fiber.Map{"rfq_id": int64(rfqID), "dismissed": false})
}

func (h *RFQHandler) GetRFQ(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
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
		if repository.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rfq not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch rfq"})
	}

	return c.JSON(fiber.Map{"rfq": rfq})
}

func (h *RFQHandler) CancelRFQ(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
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
	userID, err := getUserIDFromHeader(c)
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
