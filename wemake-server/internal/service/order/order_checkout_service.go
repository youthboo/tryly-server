package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/yourusername/wemake/internal/helper"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/yourusername/wemake/internal/domain"
	orderrepo "github.com/yourusername/wemake/internal/repository/order"
)

type BulkCheckoutItemInput struct {
	QuotationID int64  `json:"quotation_id"`
	AddressID   int64  `json:"address_id"`
	PaymentType string `json:"payment_type"`
}

type BulkCheckoutInput struct {
	RFQID          int64                   `json:"rfq_id"`
	UserID         int64                   `json:"-"`
	Items          []BulkCheckoutItemInput `json:"items"`
	IdempotencyKey string                  `json:"idempotency_key"`
}

type BulkCheckoutSummary struct {
	OrderCount   int     `json:"order_count"`
	TotalAmount  float64 `json:"total_amount"`
	TotalDeposit float64 `json:"total_deposit"`
}

type BulkCheckoutResult struct {
	RFQID     int64               `json:"rfq_id"`
	RFQStatus string              `json:"rfq_status"`
	Orders    []domain.Order      `json:"orders"`
	Summary   BulkCheckoutSummary `json:"summary"`
}

// CreateFromQuotation accepts a pending (PD) quotation or continues from an already-accepted (AC) quote.
// For PD: rejects sibling PD quotations, sets this quote to AC, closes the RFQ (OP→CL), then creates an order in payment-pending state.
func (s *OrderService) CreateFromQuotation(quotationID, userID int64) (*domain.Order, error) {
	var src *orderrepo.QuotationOrderSource
	var order *domain.Order
	var total float64
	var deposit float64
	var now time.Time
	if err := helper.WithTx(context.Background(), s.db, func(tx *sqlx.Tx) error {
		var err error
		src, err = s.repo.GetOrderSourceByQuotationIDTx(tx, quotationID, userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				tx.Rollback()
			}
			return err
		}
		switch src.Status {
		case "RJ":
			return ErrQuotationRejected
		case "PD":
			// Multi-factory: ยอมรับเฉพาะ quote นี้ ไม่ reject quotations อื่น
			// ไม่ปิด RFQ อัตโนมัติ — ลูกค้าสามารถยอมรับโรงงานอื่นได้ต่อ
			if err := s.quotations.UpdateStatusTx(tx, quotationID, "AC"); err != nil {
				return err
			}
		case "AC":
			// AC แล้ว — ไม่ทำอะไรกับ quotation อื่น แค่ตรวจสอบว่า order ยังไม่มี
		default:
			return ErrQuotationInvalidState
		}

		exists, err := s.repo.OrderExistsForQuoteTx(tx, quotationID)
		if err != nil {
			return err
		}
		if exists {
			return ErrOrderAlreadyExistsForQuote
		}

		// Use grand_total (VAT + commission inclusive) from the quotation.
		// Fall back to legacy formula only when grand_total was not yet calculated (old quotations).
		total = src.GrandTotal
		if total <= 0 {
			total = helper.RoundCurrency((src.PricePerPiece * float64(src.Quantity)) + src.MoldCost)
		}
		if total < 0 {
			return errors.New("invalid order total")
		}
		// Platform policy: 100% upfront payment — deposit equals full amount.
		deposit = total
		status := "PP"
		if total == 0 {
			deposit = 0
			status = "PE"
		}

		shippingDays := getShippingDays(s.db)
		now = time.Now()
		deliveryDate := calculateEstimatedDelivery(now, src.LeadTimeDays, shippingDays, src.DeliveryDate)
		order = &domain.Order{
			QuotationID:       src.QuotationID,
			UserID:            src.UserID,
			FactoryID:         src.FactoryID,
			TotalAmount:       total,
			DepositAmount:     deposit,
			Status:            status,
			EstimatedDelivery: &deliveryDate,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		if err := s.repo.CreateTx(tx, order); err != nil {
			return err
		}
		if s.schedules != nil {
			depositDueDate := deriveDefaultDepositScheduleDate(order.CreatedAt)
			if err := s.schedules.CreateTx(tx, &domain.PaymentSchedule{
				OrderID:       order.OrderID,
				InstallmentNo: 1,
				DueDate:       depositDueDate,
				Amount:        deposit,
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Activity log is best-effort; order creation should not fail if audit table/schema lags behind.
	uid := userID
	_ = s.repo.InsertActivity(order.OrderID, &uid, "ORDER_CREATED", map[string]interface{}{
		"status":         order.Status,
		"quote_id":       order.QuotationID,
		"amount":         total,
		"deposit_amount": deposit,
	})
	createNotificationSafe(s.notifications, &domain.Notification{
		UserID:  order.FactoryID,
		Type:    "ORDER_PLACED",
		Title:   "คำสั่งซื้อใหม่",
		Message: fmt.Sprintf("ลูกค้าสั่งซื้อ Order #%d", order.OrderID),
		LinkTo:  orderLink(order.OrderID),
		Data: notificationData(map[string]interface{}{
			"order_id": order.OrderID,
			"quote_id": order.QuotationID,
			"url":      orderLink(order.OrderID),
		}),
		ReferenceID: &order.OrderID,
		CreatedAt:   now,
	})
	s.notifyAcceptedQuotationInChat(src.RFQID, src.UserID, src.FactoryID, order.OrderID)
	return order, nil
}

func (s *OrderService) BulkCheckout(input BulkCheckoutInput) (*BulkCheckoutResult, error) {
	if input.RFQID <= 0 || input.UserID <= 0 || len(input.Items) == 0 {
		return nil, ErrInvalidQuotationSet
	}
	quoteIDs := make([]int64, 0, len(input.Items))
	seen := make(map[int64]struct{}, len(input.Items))
	paymentTypes := make(map[int64]string, len(input.Items))
	addressIDs := make(map[int64]struct{})
	for _, item := range input.Items {
		if item.QuotationID <= 0 {
			return nil, ErrInvalidQuotationSet
		}
		if item.AddressID <= 0 {
			return nil, ErrInvalidQuotationSet
		}
		if _, ok := seen[item.QuotationID]; ok {
			return nil, ErrInvalidQuotationSet
		}
		seen[item.QuotationID] = struct{}{}
		pt := strings.TrimSpace(strings.ToUpper(item.PaymentType))
		switch pt {
		case "", "FULL", "FP":
			pt = "FP"
		case "DEPOSIT", "DP":
			pt = "DP"
		default:
			return nil, ErrPaymentTypeInvalid
		}
		paymentTypes[item.QuotationID] = pt
		quoteIDs = append(quoteIDs, item.QuotationID)
		addressIDs[item.AddressID] = struct{}{}
	}

	now := time.Now()
	orders := make([]domain.Order, 0, len(quoteIDs))
	if err := helper.WithTx(context.Background(), s.db, func(tx *sqlx.Tx) error {
		type lockedRFQ struct {
			RFQID  int64  `db:"rfq_id"`
			UserID int64  `db:"user_id"`
			Status string `db:"status"`
		}
		var rfq lockedRFQ
		if err := tx.Get(&rfq, `
		SELECT rfq_id, user_id, status
		FROM rfqs
		WHERE rfq_id = $1
		FOR UPDATE
	`, input.RFQID); err != nil {
			return err
		}
		if rfq.UserID != input.UserID {
			return ErrNotQuotationParty
		}
		if rfq.Status != "OP" && rfq.Status != "IR" {
			return ErrRFQLocked
		}
		if len(addressIDs) > 0 {
			ids := make([]int64, 0, len(addressIDs))
			for id := range addressIDs {
				ids = append(ids, id)
			}
			var ownedCount int
			if err := tx.Get(&ownedCount, `
			SELECT COUNT(*)
			FROM addresses
			WHERE user_id = $1
			  AND address_id = ANY($2)
		`, input.UserID, pq.Array(ids)); err != nil {
				return err
			}
			if ownedCount != len(ids) {
				return ErrInvalidQuotationSet
			}
		}

		type lockedQuotation struct {
			QuoteID       int64      `db:"quote_id"`
			RFQID         int64      `db:"rfq_id"`
			FactoryID     int64      `db:"factory_id"`
			PricePerPiece float64    `db:"price_per_piece"`
			Quantity      int64      `db:"quantity"`
			MoldCost      float64    `db:"mold_cost"`
			LeadTimeDays  int64      `db:"lead_time_days"`
			DeliveryDate  *time.Time `db:"delivery_date"`
			Status        string     `db:"status"`
			GrandTotal    float64    `db:"grand_total"`
			ValidUntil    *time.Time `db:"valid_until"`
		}
		var quotes []lockedQuotation
		if err := tx.Select(&quotes, `
		SELECT q.quote_id, q.rfq_id, q.factory_id, q.price_per_piece, r.quantity,
		       q.mold_cost, q.lead_time_days, NULL::date AS delivery_date, q.status, COALESCE(q.grand_total, 0) AS grand_total,
		       COALESCE(q.valid_until, (q.create_time + (q.validity_days::text || ' day')::interval)::date) AS valid_until
		FROM quotations q
		INNER JOIN rfqs r ON r.rfq_id = q.rfq_id
		WHERE q.rfq_id = $1
		  AND q.quote_id = ANY($2)
		FOR UPDATE OF q
	`, input.RFQID, pq.Array(quoteIDs)); err != nil {
			return err
		}
		if len(quotes) != len(quoteIDs) {
			return ErrInvalidQuotationSet
		}

		shippingDays := getShippingDays(s.db)
		for _, q := range quotes {
			if q.FactoryID == input.UserID {
				return ErrSelfTransaction
			}
			if q.Status != "PD" {
				return ErrQuotationInvalidState
			}
			if q.ValidUntil != nil && q.ValidUntil.Before(now.UTC()) {
				return ErrQuotationExpired
			}
			exists, err := s.repo.OrderExistsForQuoteTx(tx, q.QuoteID)
			if err != nil {
				return err
			}
			if exists {
				return ErrOrderAlreadyExistsForQuote
			}
			total := q.GrandTotal
			if total <= 0 {
				total = helper.RoundCurrency((q.PricePerPiece * float64(q.Quantity)) + q.MoldCost)
			}
			if total < 0 {
				return errors.New("invalid order total")
			}
			deposit := total
			status := "PP"
			if total == 0 {
				status = "PE"
				deposit = 0
			}
			deliveryDate := calculateEstimatedDelivery(now, q.LeadTimeDays, shippingDays, q.DeliveryDate)
			order := &domain.Order{
				QuotationID:       q.QuoteID,
				UserID:            input.UserID,
				FactoryID:         q.FactoryID,
				TotalAmount:       total,
				DepositAmount:     deposit,
				Status:            status,
				EstimatedDelivery: &deliveryDate,
				CreatedAt:         now,
				UpdatedAt:         now,
			}
			if err := s.repo.CreateTx(tx, order); err != nil {
				return err
			}
			if _, err := tx.Exec(`UPDATE orders SET payment_type = $1 WHERE order_id = $2`, paymentTypes[q.QuoteID], order.OrderID); err != nil {
				return err
			}
			if s.schedules != nil && deposit > 0 {
				if err := s.schedules.CreateTx(tx, &domain.PaymentSchedule{
					OrderID:       order.OrderID,
					InstallmentNo: 1,
					DueDate:       deriveDefaultDepositScheduleDate(order.CreatedAt),
					Amount:        deposit,
				}); err != nil {
					return err
				}
			}
			orders = append(orders, *order)
		}

		if _, err := tx.Exec(`
		UPDATE quotations
		SET status = 'AC', is_locked = TRUE, log_timestamp = NOW()
		WHERE rfq_id = $1 AND quote_id = ANY($2)
	`, input.RFQID, pq.Array(quoteIDs)); err != nil {
			return err
		}
		return s.rfqs.MarkInReviewTx(tx, input.RFQID, input.UserID)
	}); err != nil {
		return nil, err
	}

	result := &BulkCheckoutResult{
		RFQID:     input.RFQID,
		RFQStatus: "IR",
		Orders:    orders,
		Summary: BulkCheckoutSummary{
			OrderCount: len(orders),
		},
	}
	uid := input.UserID
	for _, order := range orders {
		result.Summary.TotalAmount += order.TotalAmount
		result.Summary.TotalDeposit += order.DepositAmount
		_ = s.repo.InsertActivity(order.OrderID, &uid, "ORDER_CREATED", map[string]interface{}{
			"status":         order.Status,
			"quote_id":       order.QuotationID,
			"amount":         order.TotalAmount,
			"deposit_amount": order.DepositAmount,
			"bulk_checkout":  true,
		})
		createNotificationSafe(s.notifications, &domain.Notification{
			UserID:  order.FactoryID,
			Type:    "ORDER_PLACED",
			Title:   "คำสั่งซื้อใหม่",
			Message: fmt.Sprintf("ลูกค้าสั่งซื้อ Order #%d", order.OrderID),
			LinkTo:  orderLink(order.OrderID),
			Data: notificationData(map[string]interface{}{
				"order_id": order.OrderID,
				"quote_id": order.QuotationID,
				"url":      orderLink(order.OrderID),
			}),
			ReferenceID: &order.OrderID,
			CreatedAt:   now,
		})
	}
	return result, nil
}

func (s *OrderService) notifyAcceptedQuotationInChat(rfqID, customerID, factoryID, orderID int64) {
	if s.messages == nil {
		return
	}
	rfq, err := s.rfqs.GetByIDAny(rfqID)
	if err != nil || rfq == nil || rfq.ConversationID == nil {
		return
	}
	content := fmt.Sprintf("ลูกค้ายืนยันใบเสนอราคาแล้ว · Order #%d ถูกสร้าง", orderID)
	_ = s.messages.AutoSendSystemMessage(context.Background(), *rfq.ConversationID, customerID, factoryID, content)
}
