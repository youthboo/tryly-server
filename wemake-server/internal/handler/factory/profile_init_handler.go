package factory

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/helper"
	authservice "github.com/yourusername/wemake/internal/service/auth"
	catalogservice "github.com/yourusername/wemake/internal/service/catalog"
	factoryservice "github.com/yourusername/wemake/internal/service/factory"
	masterservice "github.com/yourusername/wemake/internal/service/master"
	userservice "github.com/yourusername/wemake/internal/service/user"
)

type ProfileInitHandler struct {
	factory *factoryservice.FactoryService
	master  *masterservice.MasterService
	catalog *catalogservice.CatalogService
	address *userservice.AddressService
	auth    *authservice.AuthService
}

func NewProfileInitHandler(
	factory *factoryservice.FactoryService,
	master *masterservice.MasterService,
	catalog *catalogservice.CatalogService,
	address *userservice.AddressService,
	auth *authservice.AuthService,
) *ProfileInitHandler {
	return &ProfileInitHandler{factory: factory, master: master, catalog: catalog, address: address, auth: auth}
}

// GET /factories/me/profile-init
// Returns all data needed to render the factory profile page in one request.
func (h *ProfileInitHandler) GetProfileInit(c *fiber.Ctx) error {
	userID, _, err := helper.RequireAPIFactoryUser(
		c, h.auth,
		helper.BadRequestAPIError("INVALID_USER_CONTEXT", "invalid user context"),
		helper.UnauthorizedAPIError("USER_NOT_FOUND", "user not found"),
		helper.ForbiddenAPIError("FACTORY_ROLE_REQUIRED", "factory role required"),
	)
	if err != nil {
		return err
	}

	// ---- parallel fetch of all static master data + factory profile ----
	type result[T any] struct {
		data T
		err  error
	}

	var (
		wg           sync.WaitGroup
		factoryRes   result[*domain.FactoryPublicDetail]
		typesRes     result[[]domain.LBIFactoryType]
		categoriesRes result[[]domain.Category]
		addressesRes result[[]domain.Address]
		certsRes     result[[]domain.LBIMasterCertificate]
	)

	wg.Add(5)

	go func() {
		defer wg.Done()
		data, err := h.factory.GetPublicDetail(userID)
		factoryRes = result[*domain.FactoryPublicDetail]{data, err}
	}()

	go func() {
		defer wg.Done()
		data, err := h.master.GetFactoryTypes()
		typesRes = result[[]domain.LBIFactoryType]{data, err}
	}()

	go func() {
		defer wg.Done()
		data, err := h.catalog.GetCategories(domain.CatalogScopeAll)
		categoriesRes = result[[]domain.Category]{data, err}
	}()

	go func() {
		defer wg.Done()
		data, err := h.address.ListByUserID(userID)
		addressesRes = result[[]domain.Address]{data, err}
	}()

	go func() {
		defer wg.Done()
		data, err := h.master.GetCertificates()
		certsRes = result[[]domain.LBIMasterCertificate]{data, err}
	}()

	wg.Wait()

	if factoryRes.err != nil {
		return helper.WriteAPIError(c, helper.InternalServerAPIError("FETCH_FACTORY_FAILED", "failed to fetch factory profile"))
	}

	// ---- fetch sub-categories for each of the factory's current categories ----
	subCategories := make([]domain.SubCategory, 0)
	if factoryRes.data != nil && len(factoryRes.data.Categories) > 0 {
		var subMu sync.Mutex
		var subWg sync.WaitGroup
		for _, cat := range factoryRes.data.Categories {
			subWg.Add(1)
			go func(catID int64) {
				defer subWg.Done()
				subs, err := h.catalog.GetSubCategories(catID)
				if err != nil || len(subs) == 0 {
					return
				}
				subMu.Lock()
				subCategories = append(subCategories, subs...)
				subMu.Unlock()
			}(int64(cat.CategoryID))
		}
		subWg.Wait()
	}

	// ---- safe defaults for nil slices ----
	factoryTypes := typesRes.data
	if factoryTypes == nil {
		factoryTypes = []domain.LBIFactoryType{}
	}
	lbiCategories := categoriesRes.data
	if lbiCategories == nil {
		lbiCategories = []domain.Category{}
	}
	addresses := addressesRes.data
	if addresses == nil {
		addresses = []domain.Address{}
	}
	certTypes := certsRes.data
	if certTypes == nil {
		certTypes = []domain.LBIMasterCertificate{}
	}

	return c.JSON(fiber.Map{
		"factory":          factoryRes.data,
		"factory_types":    factoryTypes,
		"lbi_categories":   lbiCategories,
		"addresses":        addresses,
		"certificate_types": certTypes,
		"sub_categories":   subCategories,
	})
}
