package quotation

import (
	"context"
	"fmt"

	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/helper"
)

type notificationCreator interface {
	Create(*domain.Notification) error
}

type systemMessageSender interface {
	AutoSendSystemMessage(context.Context, int64, int64, int64, string) error
	AutoSendQuotationCard(context.Context, int64, int64, *domain.Quotation) error
}

var WithTx = helper.WithTx

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
