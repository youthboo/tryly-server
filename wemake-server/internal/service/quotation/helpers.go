package quotation

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/domain"
)

type notificationCreator interface {
	Create(*domain.Notification) error
}

type systemMessageSender interface {
	AutoSendSystemMessage(context.Context, int64, int64, int64, string) error
	AutoSendQuotationCard(context.Context, int64, int64, *domain.Quotation) error
}

func WithTx(ctx context.Context, db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func notificationData(payload map[string]interface{}) *domain.JSONB {
	if len(payload) == 0 {
		return nil
	}
	data := domain.JSONB(payload)
	return &data
}

func createNotificationSafe(s notificationCreator, noti *domain.Notification) {
	if s == nil || noti == nil || noti.UserID <= 0 {
		return
	}
	_ = s.Create(noti)
}

func orderLink(orderID int64) string {
	return fmt.Sprintf("/orders/%d", orderID)
}

func quoteLink(quoteID int64) string {
	return fmt.Sprintf("/quotations/%d", quoteID)
}
