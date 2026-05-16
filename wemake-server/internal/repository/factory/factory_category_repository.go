package factory

import (
	"database/sql"
	"errors"
	"github.com/yourusername/wemake/internal/helper"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/yourusername/wemake/internal/domain"
)

func (r *FactoryRepository) ListFactoryCategories(factoryID int64) ([]domain.FactoryProfileCategory, error) {
	return r.selectFactoryCategories(factoryID)
}

func (r *FactoryRepository) AddFactoryCategory(factoryID, categoryID int64) error {
	var dup bool
	if err := r.db.Get(&dup, `
		SELECT EXISTS(
			SELECT 1 FROM map_factory_categories WHERE factory_id = $1 AND category_id = $2
		)`, factoryID, categoryID); err != nil {
		return err
	}
	if dup {
		return ErrDuplicateFactoryCategory
	}
	_, err := r.db.Exec(
		`INSERT INTO map_factory_categories (factory_id, category_id) VALUES ($1, $2)`,
		factoryID, categoryID,
	)
	if err == nil {
		return nil
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return ErrDuplicateFactoryCategory
		case "23503":
			return ErrInvalidFactoryCategory
		}
	}
	return err
}

func (r *FactoryRepository) RemoveFactoryCategory(factoryID, categoryID int64) error {
	res, err := r.db.Exec(
		`DELETE FROM map_factory_categories WHERE factory_id = $1 AND category_id = $2`,
		factoryID, categoryID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *FactoryRepository) FactoryExistsActive(factoryID int64) (bool, error) {
	return r.factoryExistsActive(factoryID)
}

func (r *FactoryRepository) ListFactorySubCategories(factoryID int64) ([]domain.FactoryProfileSubCategory, error) {
	return r.selectFactorySubCategories(factoryID)
}

func (r *FactoryRepository) AddFactorySubCategory(factoryID, subCategoryID int64) error {
	var dup bool
	if err := r.db.Get(&dup, `
		SELECT EXISTS(
			SELECT 1 FROM map_factory_sub_categories WHERE factory_id = $1 AND sub_category_id = $2
		)`, factoryID, subCategoryID); err != nil {
		return err
	}
	if dup {
		return ErrDuplicateFactorySubCategory
	}
	_, err := r.db.Exec(
		`INSERT INTO map_factory_sub_categories (factory_id, sub_category_id) VALUES ($1, $2)`,
		factoryID, subCategoryID,
	)
	if err == nil {
		return nil
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return ErrDuplicateFactorySubCategory
		case "23503":
			return ErrInvalidFactorySubCategory
		}
	}
	return err
}

func (r *FactoryRepository) RemoveFactorySubCategory(factoryID, subCategoryID int64) error {
	res, err := r.db.Exec(
		`DELETE FROM map_factory_sub_categories WHERE factory_id = $1 AND sub_category_id = $2`,
		factoryID, subCategoryID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *FactoryRepository) ReplaceFactoryCategories(factoryID int64, categoryIDs []int64) error {
	return helper.WithTx(nil, r.db, func(tx *sqlx.Tx) error {
		if _, err := tx.Exec(`DELETE FROM map_factory_categories WHERE factory_id = $1`, factoryID); err != nil {
			return err
		}
		for _, categoryID := range categoryIDs {
			if _, err := tx.Exec(
				`INSERT INTO map_factory_categories (factory_id, category_id) VALUES ($1, $2)`,
				factoryID, categoryID,
			); err != nil {
				var pqErr *pq.Error
				if errors.As(err, &pqErr) && pqErr.Code == "23503" {
					return ErrInvalidFactoryCategory
				}
				return err
			}
		}
		return nil
	})
}

func (r *FactoryRepository) ReplaceFactorySubCategories(factoryID int64, subCategoryIDs []int64) error {
	return helper.WithTx(nil, r.db, func(tx *sqlx.Tx) error {
		if _, err := tx.Exec(`DELETE FROM map_factory_sub_categories WHERE factory_id = $1`, factoryID); err != nil {
			return err
		}
		for _, subCategoryID := range subCategoryIDs {
			if _, err := tx.Exec(
				`INSERT INTO map_factory_sub_categories (factory_id, sub_category_id) VALUES ($1, $2)`,
				factoryID, subCategoryID,
			); err != nil {
				var pqErr *pq.Error
				if errors.As(err, &pqErr) && pqErr.Code == "23503" {
					return ErrInvalidFactorySubCategory
				}
				return err
			}
		}
		return nil
	})
}
