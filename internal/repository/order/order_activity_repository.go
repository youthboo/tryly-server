package order

import (
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/domain"
)

func (r *OrderRepository) InsertActivity(orderID int64, actorUserID *int64, eventCode string, payload map[string]interface{}) error {
	var payloadArg interface{}
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		payloadArg = b
	}
	_, err := r.db.Exec(
		`INSERT INTO order_activity_log (order_id, actor_user_id, event_code, payload) VALUES ($1, $2, $3, $4)`,
		orderID, actorUserID, eventCode, payloadArg,
	)
	return err
}

// InsertActivityTx writes an order activity row inside an existing transaction.
func (r *OrderRepository) InsertActivityTx(tx *sqlx.Tx, orderID int64, actorUserID *int64, eventCode string, payload map[string]interface{}) error {
	var payloadArg interface{}
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		payloadArg = b
	}
	_, err := tx.Exec(
		`INSERT INTO order_activity_log (order_id, actor_user_id, event_code, payload) VALUES ($1, $2, $3, $4)`,
		orderID, actorUserID, eventCode, payloadArg,
	)
	return err
}

func (r *OrderRepository) ListActivity(orderID int64) ([]domain.OrderActivityEntry, error) {
	var rows []domain.OrderActivityEntry
	err := r.db.Select(&rows, `
		SELECT activity_id, order_id, actor_user_id, event_code, payload, created_at
		FROM order_activity_log
		WHERE order_id = $1
		ORDER BY created_at ASC
	`, orderID)
	return rows, err
}
