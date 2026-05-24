package catalog

import (
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/logger"
	catalogrepo "github.com/yourusername/wemake/internal/repository/catalog"
	showcaseservice "github.com/yourusername/wemake/internal/service/showcase"
)

var exploreShowcaseTypes = []string{"PD", "MT", "PM", "ID"}

type CatalogService struct {
	repo            *catalogrepo.CatalogRepository
	showcaseService *showcaseservice.ShowcaseService
}

func NewCatalogService(repo *catalogrepo.CatalogRepository, showcaseService *showcaseservice.ShowcaseService) *CatalogService {
	return &CatalogService{repo: repo, showcaseService: showcaseService}
}

func (s *CatalogService) GetExplore() (*domain.ExploreResponse, error) {
	logger.Debug("building explore page data")

	categories, err := s.GetCategories(domain.CatalogScopeAll, 6)
	if err != nil {
		return nil, err
	}
	if categories == nil {
		categories = []domain.Category{}
	}

	showcases, err := s.showcaseService.GetHomeShowcases(exploreShowcaseTypes, 8)
	if err != nil {
		return nil, err
	}

	return &domain.ExploreResponse{
		Categories: categories,
		Showcases:  showcases,
	}, nil
}

func (s *CatalogService) GetCategories(scope string, limit int) ([]domain.Category, error) {
	return s.repo.GetCategories(scope, limit)
}

func (s *CatalogService) GetSubCategories(categoryID int64) ([]domain.SubCategory, error) {
	return s.repo.GetSubCategories(categoryID)
}

func (s *CatalogService) GetAllSubCategories(scope string) ([]domain.SubCategory, error) {
	return s.repo.GetAllSubCategories(scope)
}

func (s *CatalogService) GetCategoriesWithSubs(scope string, limit int) ([]domain.CategoryWithSubs, error) {
	return s.repo.GetCategoriesWithSubs(scope, limit)
}

func (s *CatalogService) GetUnits() ([]domain.Unit, error) {
	return s.repo.GetUnits()
}
