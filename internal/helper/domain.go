package helper

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/domain"
)

func NotificationData(payload map[string]interface{}) *domain.JSONB {
	if len(payload) == 0 {
		return nil
	}
	data := domain.JSONB(payload)
	return &data
}

func CreateNotificationSafe(s interface {
	Create(*domain.Notification) error
}, noti *domain.Notification) {
	if s == nil || noti == nil || noti.UserID <= 0 {
		return
	}
	_ = s.Create(noti)
}

func InsertDomainEventTx(tx *sqlx.Tx, eventType string, payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO domain_events (event_type, payload) VALUES ($1, $2)`, eventType, b)
	return err
}

func TrimNotificationPreview(value string, max int) string {
	value = strings.TrimSpace(value)
	if value == "" || max <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max]) + "..."
}

func OrderLink(orderID int64) string {
	return fmt.Sprintf("/orders/%d", orderID)
}

func RFQLink(rfqID int64) string     { return fmt.Sprintf("/rfqs/%d", rfqID) }
func FactoryRFQLink(rfqID int64) string { return fmt.Sprintf("/factory/rfqs/%d", rfqID) }
func FactoryOrderLink(orderID int64) string { return fmt.Sprintf("/factory/orders/%d", orderID) }
