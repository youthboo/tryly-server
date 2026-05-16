package repository

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/domain"
)

func (r *ShowcaseRepository) GetSections(showcaseID, factoryID int64) ([]domain.ShowcaseSection, error) {
	// Verify ownership
	var ownerID int64
	if err := r.db.Get(&ownerID, `SELECT factory_id FROM factory_showcases WHERE showcase_id = $1`, showcaseID); err != nil {
		return nil, sql.ErrNoRows
	}
	if ownerID != factoryID {
		return nil, domain.ErrForbidden
	}

	var rows []sectionRow
	if err := r.db.Select(&rows, `
		SELECT
			s.section_id, s.section_type, s.section_title, s.sort_order,
			i.item_id,
			i.title       AS item_title,
			i.description AS item_description,
			i.icon_name,
			i.sort_order  AS item_sort_order
		FROM showcase_sections s
		LEFT JOIN showcase_section_items i ON s.section_id = i.section_id
		WHERE s.showcase_id = $1
		ORDER BY s.sort_order, s.section_id, i.sort_order, i.item_id
	`, showcaseID); err != nil {
		return nil, err
	}
	return aggregateSections(rows), nil
}

func (r *ShowcaseRepository) BulkReplaceSections(showcaseID, factoryID int64, inputs []domain.ShowcaseSectionInput) error {
	// Verify ownership
	var ownerID int64
	if err := r.db.Get(&ownerID, `SELECT factory_id FROM factory_showcases WHERE showcase_id = $1`, showcaseID); err != nil {
		return sql.ErrNoRows
	}
	if ownerID != factoryID {
		return domain.ErrForbidden
	}

	return withTx(nil, r.db, func(tx *sqlx.Tx) error {
		if _, err := tx.Exec(`DELETE FROM showcase_sections WHERE showcase_id = $1`, showcaseID); err != nil {
			return err
		}

		for _, sec := range inputs {
			var sectionID int64
			if err := tx.QueryRow(
				`INSERT INTO showcase_sections (showcase_id, section_type, section_title, sort_order)
				 VALUES ($1, $2, $3, $4) RETURNING section_id`,
				showcaseID, sec.SectionType, sec.SectionTitle, sec.SortOrder,
			).Scan(&sectionID); err != nil {
				return err
			}
			for _, item := range sec.Items {
				if _, err := tx.Exec(
					`INSERT INTO showcase_section_items (section_id, title, description, icon_name, sort_order)
					 VALUES ($1, $2, $3, $4, $5)`,
					sectionID, item.Title, item.Description, item.IconName, item.SortOrder,
				); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *ShowcaseRepository) GetSpecs(showcaseID, factoryID int64) ([]domain.ShowcaseSpec, error) {
	var ownerID int64
	if err := r.db.Get(&ownerID, `SELECT factory_id FROM factory_showcases WHERE showcase_id = $1`, showcaseID); err != nil {
		return nil, sql.ErrNoRows
	}
	if ownerID != factoryID {
		return nil, domain.ErrForbidden
	}
	var specs []domain.ShowcaseSpec
	if err := r.db.Select(&specs, `
		SELECT spec_id, showcase_id, spec_key, spec_value, sort_order
		FROM showcase_specs
		WHERE showcase_id = $1
		ORDER BY sort_order, spec_id
	`, showcaseID); err != nil {
		return nil, err
	}
	if specs == nil {
		specs = []domain.ShowcaseSpec{}
	}
	return specs, nil
}

func (r *ShowcaseRepository) BulkReplaceSpecs(showcaseID, factoryID int64, inputs []domain.ShowcaseSpecInput) error {
	var ownerID int64
	if err := r.db.Get(&ownerID, `SELECT factory_id FROM factory_showcases WHERE showcase_id = $1`, showcaseID); err != nil {
		return sql.ErrNoRows
	}
	if ownerID != factoryID {
		return domain.ErrForbidden
	}

	return withTx(nil, r.db, func(tx *sqlx.Tx) error {
		if _, err := tx.Exec(`DELETE FROM showcase_specs WHERE showcase_id = $1`, showcaseID); err != nil {
			return err
		}
		for _, spec := range inputs {
			if _, err := tx.Exec(
				`INSERT INTO showcase_specs (showcase_id, spec_key, spec_value, sort_order)
				 VALUES ($1, $2, $3, $4)`,
				showcaseID, spec.SpecKey, spec.SpecValue, spec.SortOrder,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ShowcaseRepository) DeleteSection(showcaseID, sectionID, factoryID int64) error {
	res, err := r.db.Exec(`
		DELETE FROM showcase_sections
		WHERE section_id = $1
		  AND showcase_id = $2
		  AND EXISTS (SELECT 1 FROM factory_showcases WHERE showcase_id = $2 AND factory_id = $3)
	`, sectionID, showcaseID, factoryID)
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
