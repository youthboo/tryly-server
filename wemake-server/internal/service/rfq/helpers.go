package rfq

import (
	"fmt"

	"github.com/yourusername/wemake/internal/domain"
)

type notificationCreator interface {
	Create(*domain.Notification) error
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

func rfqLink(rfqID int64) string {
	return fmt.Sprintf("/rfqs/%d", rfqID)
}
