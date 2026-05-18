package catalog

import (
	"github.com/yourusername/wemake/internal/domain"
	catalogrepo "github.com/yourusername/wemake/internal/repository/catalog"
)

type CatalogService struct {
	repo *catalogrepo.CatalogRepository
}

func NewCatalogService(repo *catalogrepo.CatalogRepository) *CatalogService {
	return &CatalogService{repo: repo}
}

func (s *CatalogService) GetCategories(scope string) ([]domain.Category, error) {
	return s.repo.GetCategories(scope)
}

func (s *CatalogService) GetSubCategories(categoryID int64) ([]domain.SubCategory, error) {
	return s.repo.GetSubCategories(categoryID)
}

func (s *CatalogService) GetAllSubCategories(scope string) ([]domain.SubCategory, error) {
	return s.repo.GetAllSubCategories(scope)
}

func (s *CatalogService) GetUnits() ([]domain.Unit, error) {
	return s.repo.GetUnits()
}
