package repository

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/domain"
)

type CatalogRepository struct {
	db *sqlx.DB
}

func NewCatalogRepository(db *sqlx.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) GetCategories(scope string) ([]domain.Category, error) {
	var categories []domain.Category
	scope = strings.TrimSpace(strings.ToUpper(scope))
	query := "SELECT category_id, name, COALESCE(scope, 'PD') AS scope FROM lbi_categories"
	args := []interface{}{}
	if scope == "" {
		scope = "PD"
	}
	if scope != "ALL" {
		query += " WHERE COALESCE(scope, 'PD') = $1"
		args = append(args, scope)
	}
	query += " ORDER BY category_id ASC"
	err := r.db.Select(&categories, query, args...)
	return categories, err
}

func (r *CatalogRepository) GetSubCategories(categoryID int64) ([]domain.SubCategory, error) {
	var subCategories []domain.SubCategory
	query := `
		SELECT sub_category_id, category_id, name, sort_order, status
		FROM lbi_sub_categories
		WHERE category_id = $1 AND status = '1'
		ORDER BY sort_order ASC, sub_category_id ASC, name ASC
	`
	err := r.db.Select(&subCategories, query, categoryID)
	return subCategories, err
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
