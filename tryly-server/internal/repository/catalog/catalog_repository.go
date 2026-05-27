package catalog

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/domainutil"
)

type CatalogRepository struct {
	db *sqlx.DB
}

func NewCatalogRepository(db *sqlx.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) GetCategories(scope string, limit int) ([]domain.Category, error) {
	var categories []domain.Category
	scope = domainutil.NormalizeStatus(scope)
	query := "SELECT category_id, name, COALESCE(scope, $1) AS scope FROM lbi_categories"
	args := []interface{}{}
	if scope == "" {
		scope = domain.CatalogScopeProduct
	}
	args = append(args, domain.CatalogScopeProduct)
	if scope != domain.CatalogScopeAll {
		query += " WHERE COALESCE(scope, $1) = $2"
		args = append(args, scope)
	}
	query += " ORDER BY CASE WHEN name = 'ทั้งหมด' THEN 0 ELSE 1 END ASC, category_id ASC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
		args = append(args, limit)
	}
	err := r.db.Select(&categories, query, args...)
	return categories, err
}

func (r *CatalogRepository) GetSubCategories(categoryID int64) ([]domain.SubCategory, error) {
	var subCategories []domain.SubCategory
	query := `
		SELECT sub_category_id, category_id, name, sort_order, status
		FROM lbi_sub_categories
		WHERE category_id = $1 AND status = '1'
		ORDER BY CASE WHEN name = 'ทั้งหมด' THEN 0 ELSE 1 END ASC, sort_order ASC, sub_category_id ASC, name ASC
	`
	err := r.db.Select(&subCategories, query, categoryID)
	return subCategories, err
}

func (r *CatalogRepository) GetAllSubCategories(scope string) ([]domain.SubCategory, error) {
	var subs []domain.SubCategory
	query := `
		SELECT s.sub_category_id, s.category_id, s.name, s.sort_order, s.status
		FROM lbi_sub_categories s
		JOIN lbi_categories c ON c.category_id = s.category_id
		WHERE s.status = '1'
	`
	args := []interface{}{}
	if scope != "" && scope != domain.CatalogScopeAll {
		query += " AND COALESCE(c.scope, $1) = $1"
		args = append(args, scope)
	}
	query += " ORDER BY s.category_id ASC, CASE WHEN s.name = 'ทั้งหมด' THEN 0 ELSE 1 END ASC, s.sort_order ASC, s.sub_category_id ASC"
	err := r.db.Select(&subs, query, args...)
	return subs, err
}

func (r *CatalogRepository) GetCategoriesWithSubs(scope string, limit int) ([]domain.CategoryWithSubs, error) {
	cats, err := r.GetCategories(scope, limit)
	if err != nil {
		return nil, err
	}
	allSubs, err := r.GetAllSubCategories(scope)
	if err != nil {
		return nil, err
	}
	subsByCategory := make(map[int64][]domain.SubCategory, len(cats))
	for _, sub := range allSubs {
		subsByCategory[sub.CategoryID] = append(subsByCategory[sub.CategoryID], sub)
	}
	result := make([]domain.CategoryWithSubs, 0, len(cats))
	for _, cat := range cats {
		subs := subsByCategory[cat.CategoryID]
		if subs == nil {
			subs = []domain.SubCategory{}
		}
		result = append(result, domain.CategoryWithSubs{
			CategoryID:    cat.CategoryID,
			Name:          cat.Name,
			Scope:         cat.Scope,
			SubCategories: subs,
		})
	}
	return result, nil
}

func (r *CatalogRepository) GetUnits() ([]domain.Unit, error) {
	var units []domain.Unit
	query := `
		SELECT unit_id, unit_name_th AS name, unit_name_en
		FROM lbi_units
		WHERE status = '1'
		ORDER BY unit_id ASC
	`
	err := r.db.Select(&units, query)
	return units, err
}
