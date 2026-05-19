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
		wg          sync.WaitGroup
		categories  []domain.Category
		showcases   map[string][]domain.ShowcaseExploreItem
		promoSlides []domain.HomePromoSlide
		catErr      error
		showErr     error
		slideErr    error
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		categories, catErr = h.catalogService.GetCategories(domain.CatalogScopeAll, 6)
	}()

	go func() {
		defer wg.Done()
		showcases, showErr = h.showcaseService.GetHomeShowcases(exploreShowcaseTypes, 8)
	}()

	go func() {
		defer wg.Done()
		promoSlides, slideErr = h.showcaseService.ListHomePromoSlides(5)
	}()

	wg.Wait()

	if catErr != nil {
		return helper.JSONInternal(c, "failed to fetch categories")
	}
	if showErr != nil {
		return helper.JSONInternal(c, "failed to fetch showcases")
	}
	if slideErr != nil {
		// promo slides are non-critical — degrade gracefully
		promoSlides = []domain.HomePromoSlide{}
	}

	if categories == nil {
		categories = []domain.Category{}
	}
	if promoSlides == nil {
		promoSlides = []domain.HomePromoSlide{}
	}

	return c.JSON(domain.ExploreResponse{
		Categories:  categories,
		Showcases:   showcases,
		PromoSlides: promoSlides,
	})
}
