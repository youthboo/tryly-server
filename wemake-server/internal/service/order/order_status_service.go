package order

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/yourusername/wemake/internal/domain"
	domainstatus "github.com/yourusername/wemake/internal/domain/status"
	"github.com/yourusername/wemake/internal/helper"
)

func (s *OrderService) UpdateStatus(orderID int64, status string, actorUserID *int64) error {
	status = domainstatus.NormalizeOrder(status)
	if err := s.repo.UpdateStatus(orderID, status); err != nil {
		return err
	}
	return s.repo.InsertActivity(orderID, actorUserID, "ORDER_STATUS", map[string]interface{}{
		"status": status,
	})
}

func (s *OrderService) Cancel(orderID, userID int64, role string) error {
	order, err := s.repo.GetByParticipant(orderID, userID, role)
	if err != nil {
		return err
	}
	if !domainstatus.IsCancellableOrder(order.Status) {
		return ErrOrderCannotBeCancelled
	}
	if err := s.repo.UpdateStatus(orderID, domain.OrderStatusCancelledByCustomer); err != nil {
		return err
	}
	if err := s.repo.InsertActivity(orderID, &userID, "ORDER_CANCELLED", map[string]interface{}{
		"status":          domain.OrderStatusCancelledByCustomer,
		"previous_status": order.Status,
	}); err != nil {
		return err
	}
	now := time.Now()
	for _, recipient := range []int64{order.UserID, order.FactoryID} {
		helper.CreateNotificationSafe(s.notifications, &domain.Notification{
			UserID:  recipient,
			Type:    "ORDER_CANCELLED",
			Title:   "คำสั่งซื้อถูกยกเลิก",
			Message: fmt.Sprintf("Order #%d ถูกยกเลิก", orderID),
			LinkTo:  helper.OrderLink(orderID),
			Data: helper.NotificationData(map[string]interface{}{
				"order_id": orderID,
				"url":      helper.OrderLink(orderID),
			}),
			ReferenceID: &orderID,
			CreatedAt:   now,
		})
	}
	return nil
}

func (s *OrderService) MarkShipped(orderID, factoryID int64, trackingNo, courier string) error {
	trackingNo = strings.TrimSpace(trackingNo)
	courier = strings.TrimSpace(courier)
	if trackingNo == "" || courier == "" {
		return ErrShipOrderInvalid
	}
	order, err := s.repo.GetByParticipant(orderID, factoryID, domain.RoleFactory)
	if err != nil {
		return err
	}
	if order.Status != domain.OrderStatusProduction &&
		order.Status != domain.OrderStatusQualityCheck &&
		order.Status != domain.OrderStatusShipping {
		return sql.ErrNoRows
	}
	if err := s.repo.MarkShipped(orderID, factoryID, trackingNo, courier); err != nil {
		return err
	}
	uid := factoryID
	if err := s.repo.InsertActivity(orderID, &uid, "ORDER_SHIPPED", map[string]interface{}{
		"status":      "SH",
		"tracking_no": trackingNo,
		"courier":     courier,
	}); err != nil {
		return err
	}
	helper.CreateNotificationSafe(s.notifications, &domain.Notification{
		UserID:  order.UserID,
		Type:    "ORDER_SHIPPED",
		Title:   "สินค้ากำลังจัดส่ง",
		Message: fmt.Sprintf("Tracking: %s", trackingNo),
		LinkTo:  helper.OrderLink(orderID),
		Data: helper.NotificationData(map[string]interface{}{
			"order_id":    orderID,
			"tracking_no": trackingNo,
			"courier":     courier,
			"url":         helper.OrderLink(orderID),
		}),
		ReferenceID: &orderID,
		CreatedAt:   time.Now(),
	})
	return nil
}

func (s *OrderService) AutoCloseShippedOrders() (int, error) {
	cutoff := time.Now().AddDate(0, 0, -20)
	candidates, err := s.repo.ListAutoCloseCandidates(cutoff)
	if err != nil {
		return 0, err
	}
	closed := 0
	for _, orderID := range candidates {
		if _, err := s.confirmReceiptTx(orderID, nil, "auto close after 20 days", nil, "AUTO_CLOSE_20_DAYS", true); err != nil {
			// Keep processing next orders; this job should be best-effort.
			continue
		}
		closed++
	}
	return closed, nil
}
