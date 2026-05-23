package catalog

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	catalogservice "github.com/yourusername/wemake/internal/service/catalog"
)

// ExploreHandler handles GET /api/v1/explore — returns categories + showcases in one shot.
type ExploreHandler struct {
	catalogService *catalogservice.CatalogService
}

func NewExploreHandler(catalogService *catalogservice.CatalogService) *ExploreHandler {
	return &ExploreHandler{catalogService: catalogService}
}

func (h *ExploreHandler) GetExplore(c *fiber.Ctx) error {
	resp, err := h.catalogService.GetExplore()
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch explore data")
	}
	return c.JSON(resp)
}
