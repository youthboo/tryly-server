package rfq

import (
	"database/sql"

	"github.com/yourusername/wemake/internal/domain"
)

func (r *RFQRepository) RFQExists(rfqID int64) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM rfqs WHERE rfq_id = $1)`, rfqID)
	return exists, err
}

func (r *RFQRepository) FactoryQuotationStatus(factoryID, rfqID int64) (string, bool, error) {
	var status string
	err := r.db.Get(&status, `
		SELECT status
		FROM quotations
		WHERE factory_id = $1 AND rfq_id = $2
		ORDER BY
			CASE status
				WHEN 'AC' THEN 1
				WHEN 'PD' THEN 2
				ELSE 3
			END,
			create_time DESC
		LIMIT 1
	`, factoryID, rfqID)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return status, true, nil
}

func (r *RFQRepository) DismissFactoryRFQ(factoryID, rfqID int64) (*domain.FactoryRFQDismissal, bool, error) {
	existing, found, err := r.GetFactoryRFQDismissal(factoryID, rfqID)
	if err != nil {
		return nil, false, err
	}
	if found {
		return existing, false, nil
	}

	item := &domain.FactoryRFQDismissal{}
	err = r.db.Get(item, `
		INSERT INTO factory_rfq_dismissals (factory_id, rfq_id)
		VALUES ($1, $2)
		RETURNING factory_id, rfq_id, dismissed_at
	`, factoryID, rfqID)
	if err != nil {
		return nil, false, err
	}
	return item, true, nil
}

func (r *RFQRepository) GetFactoryRFQDismissal(factoryID, rfqID int64) (*domain.FactoryRFQDismissal, bool, error) {
	item := &domain.FactoryRFQDismissal{}
	err := r.db.Get(item, `
		SELECT factory_id, rfq_id, dismissed_at
		FROM factory_rfq_dismissals
		WHERE factory_id = $1 AND rfq_id = $2
	`, factoryID, rfqID)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return item, true, nil
}

func (r *RFQRepository) UndismissFactoryRFQ(factoryID, rfqID int64) error {
	_, err := r.db.Exec(`
		DELETE FROM factory_rfq_dismissals
		WHERE factory_id = $1 AND rfq_id = $2
	`, factoryID, rfqID)
	return err
}
