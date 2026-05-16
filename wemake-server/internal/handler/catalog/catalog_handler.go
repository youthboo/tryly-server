package catalog

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/helper"
	catalogservice "github.com/yourusername/wemake/internal/service/catalog"
)

type CatalogHandler struct {
	service *catalogservice.CatalogService
}

func NewCatalogHandler(service *catalogservice.CatalogService) *CatalogHandler {
	return &CatalogHandler{service: service}
}

func (h *CatalogHandler) GetCategories(c *fiber.Ctx) error {
	items, err := h.service.GetCategories(c.Query("scope"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch categories"})
	}
	return c.JSON(items)
}

func (h *CatalogHandler) GetLBICategories(c *fiber.Ctx) error {
	scope := strings.TrimSpace(strings.ToUpper(c.Query("scope")))
	if scope == "" {
		scope = domain.CatalogScopeProduct
	}
	if scope != domain.CatalogScopeProduct &&
		scope != domain.CatalogScopeMaterial &&
		scope != domain.CatalogScopeAll {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "INVALID_SCOPE"})
	}
	items, err := h.service.GetCategories(scope)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch categories"})
	}
	return c.JSON(fiber.Map{"categories": items})
}

func (h *CatalogHandler) GetSubCategories(c *fiber.Ctx) error {
	categoryID, err := helper.ParsePositiveInt64Param(c, "id")
	if err != nil || categoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid category id"})
	}

	items, err := h.service.GetSubCategories(categoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch sub-categories"})
	}
	return c.JSON(items)
}

func (h *CatalogHandler) GetUnits(c *fiber.Ctx) error {
	items, err := h.service.GetUnits()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch units"})
	}
	return c.JSON(items)
}
