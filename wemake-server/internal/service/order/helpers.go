package order

import (
	"context"
	"fmt"
	"strings"

	"github.com/yourusername/wemake/internal/domain"
)

type notificationCreator interface {
	Create(*domain.Notification) error
}

type systemMessageSender interface {
	AutoSendSystemMessage(context.Context, int64, int64, int64, string) error
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

func trimNotificationPreview(value string, max int) string {
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

func orderLink(orderID int64) string {
	return fmt.Sprintf("/orders/%d", orderID)
}
