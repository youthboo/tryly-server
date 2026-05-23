package order

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/domain"
)

func (r *OrderRepository) MarkShipped(orderID, factoryID int64, trackingNo, courier string) error {
	res, err := r.db.Exec(`
		UPDATE orders
		SET status = 'SH',
		    tracking_no = $1,
		    courier = $2,
		    updated_at = NOW()
		WHERE order_id = $3 AND factory_id = $4
	`, trackingNo, courier, orderID, factoryID)
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

func (r *OrderRepository) GetByIDForUpdateTx(tx *sqlx.Tx, orderID int64) (*domain.Order, error) {
	var order domain.Order
	err := tx.Get(&order, `
		SELECT order_id, quote_id, customer_id AS user_id, factory_id, total_amount, deposit_amount, status,
		       estimated_delivery, tracking_no, courier, NULL::timestamp AS shipped_at, created_at, updated_at
		FROM orders
		WHERE order_id = $1
		FOR UPDATE
	`, orderID)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) MarkCompletedTx(tx *sqlx.Tx, orderID int64, completedAt time.Time) error {
	res, err := tx.Exec(`
		UPDATE orders
		SET status = 'CP',
		    completed_at = $2,
		    updated_at = NOW()
		WHERE order_id = $1
	`, orderID, completedAt)
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

func (r *OrderRepository) UpsertCompletedStepTx(tx *sqlx.Tx, orderID int64, actorUserID *int64, note string, completedAt time.Time) error {
	_, err := tx.Exec(`
		INSERT INTO production_updates (
			order_id, step_id, status, description, image_urls, completed_at, updated_by_user_id, created_at
		) VALUES (
			$1, 6, 'CD', $2, '[]', $3, $4, NOW()
		)
		ON CONFLICT (order_id, step_id) DO UPDATE SET
			status = 'CD',
			description = EXCLUDED.description,
			completed_at = EXCLUDED.completed_at,
			rejected_reason = NULL,
			updated_by_user_id = EXCLUDED.updated_by_user_id
	`, orderID, note, completedAt, actorUserID)
	return err
}

func (r *OrderRepository) ListAutoCloseCandidates(cutoff time.Time) ([]int64, error) {
	var ids []int64
	err := r.db.Select(&ids, `
		SELECT o.order_id
		FROM orders o
		WHERE o.status = 'SH'
		  AND o.updated_at <= $1
		  AND NOT EXISTS (
			SELECT 1
			FROM disputes d
			WHERE d.order_id = o.order_id
			  AND d.status = 'OP'
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM production_updates pu
			WHERE pu.order_id = o.order_id
			  AND pu.status = 'RJ'
		  )
		ORDER BY o.updated_at ASC
	`, cutoff)
	return ids, err
}
