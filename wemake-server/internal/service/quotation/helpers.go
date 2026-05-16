package quotation

import (
	"context"

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

func notificationData(payload map[string]interface{}) *domain.JSONB {
	return helper.NotificationData(payload)
}

func createNotificationSafe(s notificationCreator, noti *domain.Notification) {
	helper.CreateNotificationSafe(s, noti)
}

func orderLink(orderID int64) string {
	return helper.OrderLink(orderID)
}

func quoteLink(quoteID int64) string {
	return helper.QuoteLink(quoteID)
}
