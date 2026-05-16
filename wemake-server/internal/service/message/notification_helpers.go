package message

import (
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/helper"
	notificationservice "github.com/yourusername/wemake/internal/service/notification"
)

func notificationData(payload map[string]interface{}) *domain.JSONB {
	return helper.NotificationData(payload)
}

func createNotificationSafe(s *notificationservice.NotificationService, noti *domain.Notification) {
	helper.CreateNotificationSafe(s, noti)
}

func trimNotificationPreview(value string, max int) string {
	return helper.TrimNotificationPreview(value, max)
}

func orderLink(orderID int64) string {
	return helper.OrderLink(orderID)
}

func quoteLink(quoteID int64) string {
	return helper.QuoteLink(quoteID)
}

func rfqLink(rfqID int64) string {
	return helper.RFQLink(rfqID)
}
