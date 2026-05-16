package production

import (
	"database/sql"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	productionservice "github.com/yourusername/wemake/internal/service/production"
)

type ProductionHandler struct {
	service *productionservice.ProductionService
}

func NewProductionHandler(service *productionservice.ProductionService) *ProductionHandler {
	return &ProductionHandler{service: service}
}

func (h *ProductionHandler) ListSteps(c *fiber.Ctx) error {
	factoryTypeID, err := helper.ParseOptionalPositiveInt64Query(c, "factory_type_id")
	if err != nil {
		return productionError(c, fiber.StatusBadRequest, "INVALID_FACTORY_TYPE_ID", "invalid factory_type_id", nil)
	}
	steps, err := h.service.ListSteps(factoryTypeID)
	if err != nil {
		return productionInternalError(c, err)
	}
	c.Set("Cache-Control", "public, max-age=3600")
	return c.JSON(fiber.Map{"steps": steps})
}

func (h *ProductionHandler) ListUpdates(c *fiber.Ctx) error {
	userID, err := requireProductionUserID(c)
	if err != nil {
		return err
	}
	orderID, err := helper.RequireInt64Param(c, "order_id")
	if err != nil {
		return err
	}
	item, err := h.service.ListByOrderID(int64(orderID), userID)
	if err != nil {
		return productionServiceError(c, err)
	}
	return c.JSON(item)
}

func (h *ProductionHandler) CreateUpdate(c *fiber.Ctx) error {
	userID, err := requireProductionUserID(c)
	if err != nil {
		return err
	}
	orderID, err := helper.RequireInt64Param(c, "order_id")
	if err != nil {
		return err
	}
	var req dto.CreateProductionUpdateRequest
	if err := helper.ParseBody(c, &req, "invalid request payload"); err != nil {
		return productionError(c, fiber.StatusBadRequest, "INVALID_PAYLOAD", "invalid request payload", nil)
	}
	description := helper.DereferenceString(req.Notes, "")
	progressPercent := helper.DereferenceInt(req.ProgressPercent, 0)
	result, err := h.service.Upsert(int64(orderID), userID, productionservice.ProductionWriteInput{
		StepID:                 req.StepID,
		Status:                 req.Status,
		Description:            description,
		ImageURLs:              req.ImageURLs,
		ConfirmPaymentTrigger:  progressPercent > 0,
		HeaderPaymentConfirmed: strings.EqualFold(helper.HeaderString(c, "X-Confirm-Payment-Trigger"), "true"),
	})
	if err != nil {
		return productionServiceError(c, err)
	}
	return c.JSON(result)
}

func (h *ProductionHandler) RejectUpdate(c *fiber.Ctx) error {
	userID, err := requireProductionUserID(c)
	if err != nil {
		return err
	}
	updateID, err := helper.RequireInt64Param(c, "update_id")
	if err != nil {
		return err
	}
	var req dto.RejectProductionUpdateRequest
	if err := helper.ParseBody(c, &req, "invalid request payload"); err != nil {
		return productionError(c, fiber.StatusBadRequest, "INVALID_PAYLOAD", "invalid request payload", nil)
	}
	item, err := h.service.Reject(int64(updateID), userID, req.Reason)
	if err != nil {
		return productionServiceError(c, err)
	}
	return c.JSON(item)
}

func requireProductionUserID(c *fiber.Ctx) (int64, error) {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return 0, productionError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}
	return userID, nil
}

func productionServiceError(c *fiber.Ctx, err error) error {
	if productionservice.IsNotFound(err) {
		return helper.MapServiceError(c, err, productionFallbackError, productionNotFoundResponses)
	}
	if rule, ok := productionservice.AsProductionRuleError(err); ok {
		return helper.MapServiceErrorFunc(c, rule, productionFallbackError, productionRuleResponses)
	}
	return productionInternalError(c, err)
}

func productionInternalError(c *fiber.Ctx, err error) error {
	return productionError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", productionservice.ExplainProductionError(err), nil)
}

func productionError(c *fiber.Ctx, status int, code, message string, details map[string]interface{}) error {
	return helper.WriteErrorResponse(c, productionErrorResponse(status, code, message, details))
}

func productionErrorResponse(status int, code, message string, details map[string]interface{}) helper.ErrorResponse {
	errBody := fiber.Map{
		"code":    code,
		"message": message,
	}
	if len(details) > 0 {
		errBody["details"] = details
	}
	return helper.ErrorBody(status, fiber.Map{"error": errBody})
}

func productionRuleResponse(status int, code, message string) helper.ErrorResponseBuilder {
	return func(err error) helper.ErrorResponse {
		var details map[string]interface{}
		if rule, ok := productionservice.AsProductionRuleError(err); ok {
			details = rule.Details
		}
		return productionErrorResponse(status, code, message, details)
	}
}

var productionFallbackError = productionErrorResponse(fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to process production request", nil)

var productionNotFoundResponses = map[error]helper.ErrorResponse{
	sql.ErrNoRows: productionErrorResponse(fiber.StatusNotFound, "NOT_FOUND", "resource not found", nil),
}

var productionRuleResponses = map[error]helper.ErrorResponseBuilder{
	productionservice.ErrProductionNotOrderFactory:        productionRuleResponse(fiber.StatusForbidden, "NOT_ORDER_FACTORY", "factory caller does not own the order"),
	productionservice.ErrProductionNotOrderCustomer:       productionRuleResponse(fiber.StatusForbidden, "NOT_ORDER_CUSTOMER", "customer caller does not own the order"),
	productionservice.ErrProductionOrderLocked:            productionRuleResponse(fiber.StatusConflict, "ORDER_STATE_INVALID", "order is locked for production updates"),
	productionservice.ErrProductionAnotherStepInProgress:  productionRuleResponse(fiber.StatusConflict, "ANOTHER_STEP_IN_PROGRESS", "another step is already in progress"),
	productionservice.ErrProductionInvalidStateTransition: productionRuleResponse(fiber.StatusConflict, "STEP_LOCKED", "invalid or locked step transition"),
	productionservice.ErrProductionDownstreamInFlight:     productionRuleResponse(fiber.StatusConflict, "DOWNSTREAM_IN_FLIGHT", "cannot reject because downstream steps are already active"),
	productionservice.ErrProductionStepOrderViolation:     productionRuleResponse(fiber.StatusConflict, "STEP_LOCKED", "previous step must be completed first"),
	productionservice.ErrProductionInsufficientEvidence:   productionRuleResponse(fiber.StatusUnprocessableEntity, "INSUFFICIENT_EVIDENCE", "insufficient evidence"),
	productionservice.ErrProductionPaymentConfirmRequired: productionRuleResponse(fiber.StatusUnprocessableEntity, "PAYMENT_CONFIRMATION_REQUIRED", "payment confirmation required"),
	productionservice.ErrProductionInvalidStep:            productionRuleResponse(fiber.StatusUnprocessableEntity, "INVALID_STEP_ID", "step_id must reference an active production step"),
	productionservice.ErrProductionStepIDRequired:         productionRuleResponse(fiber.StatusUnprocessableEntity, "STEP_ID_REQUIRED", "step_id is required"),
	productionservice.ErrProductionInvalidStatus:          productionRuleResponse(fiber.StatusUnprocessableEntity, "INVALID_STATUS", "status must be IP or CD"),
	productionservice.ErrProductionMaxImages:              productionRuleResponse(fiber.StatusUnprocessableEntity, "MAX_5_IMAGES", "image_urls can contain at most 5 items"),
	productionservice.ErrProductionInvalidImageFormat:     productionRuleResponse(fiber.StatusUnprocessableEntity, "INVALID_IMAGE_FORMAT", "image_urls must be a non-empty array of unique URL strings"),
	productionservice.ErrProductionInvalidImageURL:        productionRuleResponse(fiber.StatusUnprocessableEntity, "INVALID_IMAGE_URL", "image_urls must be unique HTTP/HTTPS URLs"),
	productionservice.ErrProductionDescriptionTooLong:     productionRuleResponse(fiber.StatusUnprocessableEntity, "INVALID_DESCRIPTION", "description must be 2000 characters or fewer"),
	productionservice.ErrProductionReasonRequired:         productionRuleResponse(fiber.StatusUnprocessableEntity, "REASON_REQUIRED", "rejected_reason is required"),
}
