package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/domain"
	domainstatus "github.com/yourusername/wemake/internal/domain/status"
	"github.com/yourusername/wemake/internal/helper"
	walletrepo "github.com/yourusername/wemake/internal/repository/wallet"
)

type ConfirmReceiptInput struct {
	Note       string
	ReceivedAt *time.Time
}

type ConfirmReceiptSettlement struct {
	FactoryUserID int64   `json:"factory_user_id"`
	WalletID      int64   `json:"wallet_id"`
	MovedAmount   float64 `json:"moved_amount"`
	PendingBefore float64 `json:"pending_before"`
	PendingAfter  float64 `json:"pending_after"`
	GoodBefore    float64 `json:"good_before"`
	GoodAfter     float64 `json:"good_after"`
}

type ConfirmReceiptResult struct {
	Success         bool                     `json:"success"`
	OrderID         int64                    `json:"order_id"`
	StatusBefore    string                   `json:"status_before"`
	StatusAfter     string                   `json:"status_after"`
	CompletedStepID int64                    `json:"completed_step_id"`
	Settlement      ConfirmReceiptSettlement `json:"settlement"`
	CompletedAt     time.Time                `json:"completed_at"`
	AlreadyComplete bool                     `json:"already_completed,omitempty"`
}

func (s *OrderService) CreatePayment(orderID, userID int64, role, paymentType string, amount float64) (*domain.Transaction, error) {
	if role != domain.RoleCustomer {
		return nil, sql.ErrNoRows
	}
	order, err := s.repo.GetByParticipant(orderID, userID, role)
	if err != nil {
		return nil, err
	}

	paymentType = strings.TrimSpace(strings.ToUpper(paymentType))
	if paymentType == "DP" {
		if err := s.ensureDepositPayable(order); err != nil {
			return nil, err
		}
	}
	expectedAmount, err := expectedPaymentAmount(order, paymentType)
	if err != nil {
		return nil, err
	}
	if amount <= 0 || amount != expectedAmount {
		return nil, ErrPaymentAmountMismatch
	}

	existing, err := s.txLedger.List(walletrepo.TransactionFilters{
		OrderID: &orderID,
		Type:    &paymentType,
	})
	if err != nil {
		return nil, err
	}
	for _, row := range existing {
		if paymentType == "DP" && row.Status == "PT" {
			return nil, ErrDepositAlreadyPaid
		}
		if row.Status != "RJ" {
			return nil, ErrPaymentAlreadyExists
		}
	}

	var item *domain.Transaction
	if err := helper.WithTx(context.Background(), s.db, func(tx *sqlx.Tx) error {
		walletID, err := s.wallets.EnsureWallet(tx, userID)
		if err != nil {
			return err
		}

		now := time.Now()
		orderIDPtr := order.OrderID
		item = &domain.Transaction{
			TxID:       "tx-" + uuid.NewString(),
			WalletID:   walletID,
			OrderID:    &orderIDPtr,
			Type:       paymentType,
			Amount:     amount,
			Status:     "ST",
			CreatedAt:  now,
			UpdatedAt:  now,
			UploadedAt: now,
		}
		if err := s.txLedger.CreateTx(tx, item); err != nil {
			return err
		}
		return s.repo.InsertActivityTx(tx, orderID, &userID, "PAYMENT_CREATED", map[string]interface{}{
			"tx_id": item.TxID, "type": paymentType, "amount": amount, "status": item.Status,
		})
	}); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *OrderService) VerifyPayment(orderID, userID int64, role, txID string) (*domain.Transaction, error) {
	order, err := s.repo.GetByParticipant(orderID, userID, role)
	if err != nil {
		return nil, err
	}

	var paymentTx *domain.Transaction
	var now time.Time
	if err := helper.WithTx(context.Background(), s.db, func(tx *sqlx.Tx) error {
		var err error
		paymentTx, err = s.txLedger.GetByIDForUpdate(tx, strings.TrimSpace(txID))
		if err != nil {
			return err
		}
		if paymentTx.OrderID == nil || *paymentTx.OrderID != orderID {
			return sql.ErrNoRows
		}
		if paymentTx.Type != "DP" && paymentTx.Type != "FP" {
			return ErrPaymentTypeInvalid
		}
		if paymentTx.Type == "DP" {
			if err := s.ensureDepositPayable(order); err != nil {
				return err
			}
		}
		if paymentTx.Status != "ST" {
			if paymentTx.Type == "DP" && paymentTx.Status == "PT" {
				return ErrDepositAlreadyPaid
			}
			return ErrPaymentStateInvalid
		}

		expectedAmount, err := expectedPaymentAmount(order, paymentTx.Type)
		if err != nil {
			return err
		}
		if paymentTx.Amount != expectedAmount {
			return ErrPaymentAmountMismatch
		}

		if _, err := s.wallets.EnsureWallet(tx, order.UserID); err != nil {
			return err
		}
		if _, err := s.wallets.EnsureWallet(tx, order.FactoryID); err != nil {
			return err
		}

		customerWallet, err := s.wallets.GetByUserIDForUpdate(tx, order.UserID)
		if err != nil {
			return err
		}
		ok, err := s.wallets.DebitGoodFund(tx, customerWallet.WalletID, paymentTx.Amount)
		if err != nil {
			return err
		}
		if !ok {
			return ErrInsufficientGoodFund
		}

		factoryWallet, err := s.wallets.GetByUserIDForUpdate(tx, order.FactoryID)
		if err != nil {
			return err
		}
		if err := s.wallets.CreditGoodFund(tx, factoryWallet.WalletID, paymentTx.Amount); err != nil {
			return err
		}

		if err := s.txLedger.PatchStatusTx(tx, paymentTx.TxID, "PT"); err != nil {
			return err
		}

		now = time.Now()
		orderIDPtr := order.OrderID
		receiveTx := &domain.Transaction{
			TxID:       "tx-" + uuid.NewString(),
			WalletID:   factoryWallet.WalletID,
			OrderID:    &orderIDPtr,
			Type:       "SC",
			Amount:     paymentTx.Amount,
			Status:     "PT",
			CreatedAt:  now,
			UpdatedAt:  now,
			UploadedAt: now,
		}
		if err := s.txLedger.CreateTx(tx, receiveTx); err != nil {
			return err
		}

		if paymentTx.Type == "DP" && (domainstatus.NormalizeOrder(order.Status) == "PP" || domainstatus.NormalizeOrder(order.Status) == "PE") {
			if err := s.repo.UpdateStatusTx(tx, orderID, "PD"); err != nil {
				return err
			}
			order.Status = "PD"

			// Recalculate estimated_delivery starting from payment date (now), not order creation date.
			// "นับจากหลังจากที่ลูกค้าจ่ายเงิน" — lead_time + shipping days counted from payment confirmation.
			shippingDays := getShippingDays(s.db)
			type quoteDelivery struct {
				LeadTimeDays int64      `db:"lead_time_days"`
				DeliveryDate *time.Time `db:"delivery_date"`
			}
			var qd quoteDelivery
			if err2 := tx.Get(&qd, `SELECT lead_time_days, NULL::date AS delivery_date FROM quotations WHERE quote_id = $1`, order.QuotationID); err2 == nil {
				est := calculateEstimatedDelivery(now, qd.LeadTimeDays, shippingDays, qd.DeliveryDate)
				if _, err2 := tx.Exec(`UPDATE orders SET estimated_delivery = $1 WHERE order_id = $2`, est, orderID); err2 != nil {
					return err2
				}
			}

			if s.schedules != nil {
				if err := s.schedules.PatchStatusByOrderAndInstallmentTx(tx, orderID, 1, "PD"); err != nil && !errors.Is(err, sql.ErrNoRows) {
					return err
				}
			}
			if err := helper.InsertDomainEventTx(tx, "order.deposit_paid", map[string]interface{}{
				"order_id": orderID,
				"tx_id":    paymentTx.TxID,
				"amount":   paymentTx.Amount,
			}); err != nil {
				return err
			}
			if err := helper.InsertDomainEventTx(tx, "cache.invalidate", map[string]interface{}{
				"paths": []string{
					fmt.Sprintf("/orders/%d", orderID),
					fmt.Sprintf("/orders/%d/production-updates", orderID),
				},
			}); err != nil {
				return err
			}
		}

		return s.repo.InsertActivityTx(tx, orderID, &userID, "PAYMENT_VERIFIED", map[string]interface{}{
			"tx_id": paymentTx.TxID, "type": paymentTx.Type, "amount": paymentTx.Amount, "status": "PT", "order_status": order.Status,
		})
	}); err != nil {
		return nil, err
	}
	paymentTx.Status = "PT"
	helper.CreateNotificationSafe(s.notifications, &domain.Notification{
		UserID:  order.FactoryID,
		Type:    "PAYMENT_RECEIVED",
		Title:   "รับชำระเงินแล้ว",
		Message: fmt.Sprintf("ได้รับการชำระเงิน ฿%.2f สำหรับ Order #%d", paymentTx.Amount, order.OrderID),
		LinkTo:  helper.OrderLink(order.OrderID),
		Data: helper.NotificationData(map[string]interface{}{
			"order_id": order.OrderID,
			"amount":   paymentTx.Amount,
			"url":      helper.OrderLink(order.OrderID),
		}),
		ReferenceID: &order.OrderID,
		CreatedAt:   now,
	})
	return paymentTx, nil
}

func (s *OrderService) ConfirmReceipt(orderID, userID int64, role string, input ConfirmReceiptInput) (*ConfirmReceiptResult, error) {
	if role != domain.RoleCustomer {
		return nil, domain.ErrForbidden
	}
	return s.confirmReceiptTx(orderID, &userID, strings.TrimSpace(input.Note), input.ReceivedAt, "CUSTOMER_CONFIRMED_RECEIPT", true)
}

func (s *OrderService) confirmReceiptTx(orderID int64, actorUserID *int64, note string, receivedAt *time.Time, activityCode string, idempotent bool) (*ConfirmReceiptResult, error) {
	var order *domain.Order
	var statusBefore string
	var completedAt time.Time
	var settlement ConfirmReceiptSettlement
	if err := helper.WithTx(context.Background(), s.db, func(tx *sqlx.Tx) error {
		var err error
		order, err = s.repo.GetByIDForUpdateTx(tx, orderID)
		if err != nil {
			return err
		}
		if actorUserID != nil && order.UserID != *actorUserID {
			return domain.ErrForbidden
		}

		statusBefore = domainstatus.NormalizeOrder(order.Status)
		if statusBefore == "CP" {
			if !idempotent {
				return ErrConfirmReceiptNotAllowed
			}
			now := time.Now()
			if receivedAt != nil {
				now = *receivedAt
			}
			completedAt = now
			return nil
		}
		if statusBefore == "CN" || statusBefore == "CC" {
			return ErrConfirmReceiptNotAllowed
		}
		if statusBefore != "SH" {
			return ErrConfirmReceiptInvalidStatus
		}

		completedAt = time.Now().UTC()
		if receivedAt != nil {
			completedAt = receivedAt.UTC()
		}

		if err := s.repo.UpsertCompletedStepTx(tx, orderID, actorUserID, note, completedAt); err != nil {
			return err
		}
		if err := s.repo.MarkCompletedTx(tx, orderID, completedAt); err != nil {
			return err
		}

		if _, err := s.wallets.EnsureWallet(tx, order.FactoryID); err != nil {
			return err
		}
		factoryWallet, err := s.wallets.GetByUserIDForUpdate(tx, order.FactoryID)
		if err != nil {
			return err
		}
		movedAmount := helper.RoundCurrency(helper.DecimalToFloat(order.TotalAmount))
		if movedAmount < 0 {
			movedAmount = 0
		}
		if err := s.wallets.MovePendingToGoodTx(tx, factoryWallet.WalletID, movedAmount); err != nil {
			return err
		}
		settlement = ConfirmReceiptSettlement{
			FactoryUserID: order.FactoryID,
			WalletID:      factoryWallet.WalletID,
			MovedAmount:   movedAmount,
			PendingBefore: helper.DecimalToFloat(factoryWallet.PendingFund),
			PendingAfter:  helper.RoundCurrency(helper.DecimalToFloat(factoryWallet.PendingFund) - movedAmount),
			GoodBefore:    helper.DecimalToFloat(factoryWallet.GoodFund),
			GoodAfter:     helper.RoundCurrency(helper.DecimalToFloat(factoryWallet.GoodFund) + movedAmount),
		}

		// Settle the factory's pending SC receivables for this order: PT -> ST.
		if err := s.txLedger.SettleFactoryReceivables(tx, order.OrderID); err != nil {
			return err
		}
		return s.repo.InsertActivityTx(tx, orderID, actorUserID, activityCode, map[string]interface{}{
			"status_before": statusBefore,
			"status_after":  "CP",
			"completed_at":  completedAt,
			"settlement":    settlement,
			"note":          note,
		})
	}); err != nil {
		return nil, err
	}
	if statusBefore == "CP" {
		return &ConfirmReceiptResult{
			Success:         true,
			OrderID:         order.OrderID,
			StatusBefore:    "CP",
			StatusAfter:     "CP",
			CompletedStepID: 6,
			CompletedAt:     completedAt,
			AlreadyComplete: true,
		}, nil
	}
	for _, recipient := range []int64{order.UserID, order.FactoryID} {
		helper.CreateNotificationSafe(s.notifications, &domain.Notification{
			UserID:  recipient,
			Type:    "ORDER_COMPLETED",
			Title:   "คำสั่งซื้อเสร็จสมบูรณ์",
			Message: fmt.Sprintf("Order #%d เสร็จสมบูรณ์", orderID),
			LinkTo:  helper.OrderLink(orderID),
			Data: helper.NotificationData(map[string]interface{}{
				"order_id": orderID,
				"url":      helper.OrderLink(orderID),
			}),
			ReferenceID: &orderID,
			CreatedAt:   completedAt,
		})
	}
	return &ConfirmReceiptResult{
		Success:         true,
		OrderID:         orderID,
		StatusBefore:    statusBefore,
		StatusAfter:     "CP",
		CompletedStepID: 6,
		Settlement:      settlement,
		CompletedAt:     completedAt,
	}, nil
}
