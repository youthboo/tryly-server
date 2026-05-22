package catalog

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/helper"
	catalogservice "github.com/yourusername/wemake/internal/service/catalog"
	showcaseservice "github.com/yourusername/wemake/internal/service/showcase"
)

// ExploreHandler handles GET /api/v1/explore — returns categories + showcases + promo slides in one shot.
type ExploreHandler struct {
	catalogService  *catalogservice.CatalogService
	showcaseService *showcaseservice.ShowcaseService
}

func NewExploreHandler(catalogService *catalogservice.CatalogService, showcaseService *showcaseservice.ShowcaseService) *ExploreHandler {
	return &ExploreHandler{catalogService: catalogService, showcaseService: showcaseService}
}

var exploreShowcaseTypes = []string{"PD", "MT", "PM", "ID"}

// GetExplore handles GET /api/v1/explore
func (h *ExploreHandler) GetExplore(c *fiber.Ctx) error {
	var (
		wg         sync.WaitGroup
		categories []domain.Category
		showcases  map[string][]domain.ShowcaseExploreItem
		catErr     error
		showErr    error
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		categories, catErr = h.catalogService.GetCategories(domain.CatalogScopeAll, 6)
	}()

	go func() {
		defer wg.Done()
		showcases, showErr = h.showcaseService.GetHomeShowcases(exploreShowcaseTypes, 8)
	}()

	wg.Wait()

	if catErr != nil {
		return helper.JSONInternal(c, "failed to fetch categories")
	}
	if showErr != nil {
		return helper.JSONInternal(c, "failed to fetch showcases")
	}

	if categories == nil {
		categories = []domain.Category{}
	}

	return c.JSON(domain.ExploreResponse{
		Categories: categories,
		Showcases:  showcases,
	})
}
