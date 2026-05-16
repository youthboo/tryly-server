package order

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
}

func notificationData(payload map[string]interface{}) *domain.JSONB {
	return helper.NotificationData(payload)
}

func createNotificationSafe(s notificationCreator, noti *domain.Notification) {
	helper.CreateNotificationSafe(s, noti)
}

func trimNotificationPreview(value string, max int) string {
	return helper.TrimNotificationPreview(value, max)
}

func orderLink(orderID int64) string {
	return helper.OrderLink(orderID)
}
