package factory

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/helper"
)

type factoryMappingTable struct {
	Table          string
	FactoryColumn  string
	RelatedColumn  string
	DuplicateError error
	InvalidError   error
}

var (
	factoryCategoryMapping = factoryMappingTable{
		Table:          "map_factory_categories",
		FactoryColumn:  "factory_id",
		RelatedColumn:  "category_id",
		DuplicateError: ErrDuplicateFactoryCategory,
		InvalidError:   ErrInvalidFactoryCategory,
	}
	factorySubCategoryMapping = factoryMappingTable{
		Table:          "map_factory_sub_categories",
		FactoryColumn:  "factory_id",
		RelatedColumn:  "sub_category_id",
		DuplicateError: ErrDuplicateFactorySubCategory,
		InvalidError:   ErrInvalidFactorySubCategory,
	}
)

func (r *FactoryRepository) ListFactoryCategories(factoryID int64) ([]domain.FactoryProfileCategory, error) {
	return r.selectFactoryCategories(factoryID)
}

func (r *FactoryRepository) AddFactoryCategory(factoryID, categoryID int64) error {
	return r.addFactoryMapping(factoryCategoryMapping, factoryID, categoryID)
}

func (r *FactoryRepository) RemoveFactoryCategory(factoryID, categoryID int64) error {
	return r.removeFactoryMapping(factoryCategoryMapping, factoryID, categoryID)
}

func (r *FactoryRepository) FactoryExistsActive(factoryID int64) (bool, error) {
	return r.factoryExistsActive(factoryID)
}

func (r *FactoryRepository) ListFactorySubCategories(factoryID int64) ([]domain.FactoryProfileSubCategory, error) {
	return r.selectFactorySubCategories(factoryID)
}

func (r *FactoryRepository) AddFactorySubCategory(factoryID, subCategoryID int64) error {
	return r.addFactoryMapping(factorySubCategoryMapping, factoryID, subCategoryID)
}

func (r *FactoryRepository) RemoveFactorySubCategory(factoryID, subCategoryID int64) error {
	return r.removeFactoryMapping(factorySubCategoryMapping, factoryID, subCategoryID)
}

func (r *FactoryRepository) ReplaceFactoryCategories(factoryID int64, categoryIDs []int64) error {
	return r.replaceFactoryMappings(factoryCategoryMapping, factoryID, categoryIDs)
}

func (r *FactoryRepository) ReplaceFactorySubCategories(factoryID int64, subCategoryIDs []int64) error {
	return r.replaceFactoryMappings(factorySubCategoryMapping, factoryID, subCategoryIDs)
}

func (r *FactoryRepository) addFactoryMapping(mapping factoryMappingTable, factoryID, relatedID int64) error {
	query := "INSERT INTO " + mapping.Table + " (" + mapping.FactoryColumn + ", " + mapping.RelatedColumn + ") VALUES ($1, $2) ON CONFLICT DO NOTHING"
	res, err := r.db.Exec(query, factoryID, relatedID)
	if err != nil {
		return mapFactoryMappingError(err, mapping)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return mapping.DuplicateError
	}
	return nil
}

func (r *FactoryRepository) removeFactoryMapping(mapping factoryMappingTable, factoryID, relatedID int64) error {
	query := "DELETE FROM " + mapping.Table + " WHERE " + mapping.FactoryColumn + " = $1 AND " + mapping.RelatedColumn + " = $2"
	res, err := r.db.Exec(query, factoryID, relatedID)
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

func (r *FactoryRepository) replaceFactoryMappings(mapping factoryMappingTable, factoryID int64, relatedIDs []int64) error {
	return helper.WithTx(nil, r.db, func(tx *sqlx.Tx) error {
		if _, err := tx.Exec("DELETE FROM "+mapping.Table+" WHERE "+mapping.FactoryColumn+" = $1", factoryID); err != nil {
			return err
		}
		for _, relatedID := range relatedIDs {
			if _, err := tx.Exec(
				"INSERT INTO "+mapping.Table+" ("+mapping.FactoryColumn+", "+mapping.RelatedColumn+") VALUES ($1, $2) ON CONFLICT DO NOTHING",
				factoryID, relatedID,
			); err != nil {
				return mapFactoryMappingError(err, mapping)
			}
		}
		return nil
	})
}

func mapFactoryMappingError(err error, mapping factoryMappingTable) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return err
	}
	switch pqErr.Code {
	case "23505":
		return mapping.DuplicateError
	case "23503":
		return mapping.InvalidError
	default:
		return err
	}
}
