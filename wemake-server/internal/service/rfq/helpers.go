package rfq

import (
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/helper"
)

type notificationCreator interface {
	Create(*domain.Notification) error
}

func notificationData(payload map[string]interface{}) *domain.JSONB {
	return helper.NotificationData(payload)
}

func createNotificationSafe(s notificationCreator, noti *domain.Notification) {
	helper.CreateNotificationSafe(s, noti)
}

func rfqLink(rfqID int64) string {
	return helper.RFQLink(rfqID)
}
