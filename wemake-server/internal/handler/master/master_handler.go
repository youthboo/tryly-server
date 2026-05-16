package master

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	masterservice "github.com/yourusername/wemake/internal/service/master"
)

type MasterHandler struct {
	service *masterservice.MasterService
}

func NewMasterHandler(service *masterservice.MasterService) *MasterHandler {
	return &MasterHandler{service: service}
}

func (h *MasterHandler) GetProvinces(c *fiber.Ctx) error {
	items, err := h.service.GetProvinces()
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_PROVINCES_FAILED", "failed to fetch provinces"))
	}
	return helper.WriteListResponse(c, items, len(items))
}

func (h *MasterHandler) GetDistricts(c *fiber.Ctx) error {
	var provinceID *int64
	if raw := c.Query("province_id"); raw != "" {
		val, err := helper.ParsePositiveInt64Value(raw, "province_id")
		if err != nil {
			return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_PROVINCE_ID", "invalid province_id"))
		}
		provinceID = &val
	}
	items, err := h.service.GetDistricts(provinceID)
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_DISTRICTS_FAILED", "failed to fetch districts"))
	}
	return helper.WriteListResponse(c, items, len(items))
}

func (h *MasterHandler) GetSubDistricts(c *fiber.Ctx) error {
	var districtID *int64
	if raw := c.Query("district_id"); raw != "" {
		val, err := helper.ParsePositiveInt64Value(raw, "district_id")
		if err != nil {
			return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_DISTRICT_ID", "invalid district_id"))
		}
		districtID = &val
	}
	items, err := h.service.GetSubDistricts(districtID)
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_SUB_DISTRICTS_FAILED", "failed to fetch sub districts"))
	}
	return helper.WriteListResponse(c, items, len(items))
}

func (h *MasterHandler) GetFactoryTypes(c *fiber.Ctx) error {
	items, err := h.service.GetFactoryTypes()
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_FACTORY_TYPES_FAILED", "failed to fetch factory types"))
	}
	return helper.WriteListResponse(c, items, len(items))
}

func (h *MasterHandler) GetProductCategories(c *fiber.Ctx) error {
	var parentID *int64
	if raw := c.Query("parent_category_id"); raw != "" {
		val, err := helper.ParsePositiveInt64Value(raw, "parent_category_id")
		if err != nil {
			return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_PARENT_CATEGORY_ID", "invalid parent_category_id"))
		}
		parentID = &val
	}
	items, err := h.service.GetProductCategories(parentID)
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_CATEGORIES_FAILED", "failed to fetch product categories"))
	}
	return helper.WriteListResponse(c, items, len(items))
}

// GetCategories is an alias for product master list (same payload as GET /master/product-categories).
func (h *MasterHandler) GetCategories(c *fiber.Ctx) error {
	items, err := h.service.GetProductCategories(nil)
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_CATEGORIES_FAILED", "failed to fetch categories"))
	}
	return helper.WriteListResponse(c, items, len(items))
}

func (h *MasterHandler) GetProductionSteps(c *fiber.Ctx) error {
	var factoryTypeID *int64
	if raw := c.Query("factory_type_id"); raw != "" {
		val, err := helper.ParsePositiveInt64Value(raw, "factory_type_id")
		if err != nil {
			return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_FACTORY_TYPE_ID", "invalid factory_type_id"))
		}
		factoryTypeID = &val
	}
	items, err := h.service.GetProductionSteps(factoryTypeID)
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_PRODUCTION_STEPS_FAILED", "failed to fetch production steps"))
	}
	return helper.WriteListResponse(c, items, len(items))
}

func (h *MasterHandler) GetUnits(c *fiber.Ctx) error {
	items, err := h.service.GetUnits()
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_UNITS_FAILED", "failed to fetch master units"))
	}
	return helper.WriteListResponse(c, items, len(items))
}

func (h *MasterHandler) GetShippingMethods(c *fiber.Ctx) error {
	items, err := h.service.GetShippingMethods()
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_SHIPPING_METHODS_FAILED", "failed to fetch shipping methods"))
	}
	return helper.WriteListResponse(c, items, len(items))
}

func (h *MasterHandler) GetCertificates(c *fiber.Ctx) error {
	items, err := h.service.GetCertificates()
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_CERTIFICATES_FAILED", "failed to fetch certificates"))
	}
	return helper.WriteListResponse(c, items, len(items))
}
