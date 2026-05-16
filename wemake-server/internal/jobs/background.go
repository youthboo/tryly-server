// Package jobs contains background goroutines that run on a timer.
// Started from main.go via jobs.Start(db).
package jobs

import (
	"encoding/json"
	log "github.com/yourusername/wemake/internal/logger"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/repository"
	"github.com/yourusername/wemake/internal/service"
)

// Start launches all background jobs. Call once from main.go after DB is ready.
// Each job runs in its own goroutine and loops forever until the process exits.
func Start(db *sqlx.DB) {
	orderService := service.NewOrderService(
		db,
		repository.NewOrderRepository(db),
		nil,
		repository.NewWalletRepository(db),
		repository.NewTransactionRepository(db),
		nil,
		nil,
		repository.NewReviewRepository(db),
		nil,
		nil,
	)
	go runExpiration(db)
	go runOrderAutoClose(orderService)
	go runMatchingNotifications(db)
}

// --------------------------------------------------------------------------
// Expiration job — runs every hour
// --------------------------------------------------------------------------

// runExpiration auto-closes overdue RFQs and quotations.
func runExpiration(db *sqlx.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Run once immediately on start, then every hour.
	expireRFQs(db)
	expireQuotations(db)
	expirePendingDeposits(db)

	for range ticker.C {
		expireRFQs(db)
		expireQuotations(db)
		expirePendingDeposits(db)
	}
}

func runOrderAutoClose(orderService *service.OrderService) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	if n, err := orderService.AutoCloseShippedOrders(); err != nil {
		log.Printf("[jobs/order-auto-close] error: %v", err)
	} else if n > 0 {
		log.Printf("[jobs/order-auto-close] auto-closed %d order(s)", n)
	}

	for range ticker.C {
		if n, err := orderService.AutoCloseShippedOrders(); err != nil {
			log.Printf("[jobs/order-auto-close] error: %v", err)
		} else if n > 0 {
			log.Printf("[jobs/order-auto-close] auto-closed %d order(s)", n)
		}
	}
}

func expireRFQs(db *sqlx.DB) {
	// RFQ expiry by deadline_date is disabled because the legacy deadline column was removed.
	// RFQ auto-close is now handled inside expireQuotations (after PD quotations expire).
}

// expireQuotations sets status = 'EX' for pending (PD) quotations where the validity window has passed.
// Expiry is determined by:
//   - valid_until column (set when factory creates/edits the quotation)
//   - fallback: create_time + COALESCE(validity_days, 7) days
//
// After expiring quotations, it also auto-closes (OP → CL) any RFQ whose
// remaining quotations are all EX/RJ (no active PD or AC left).
func expireQuotations(db *sqlx.DB) {
	// ขั้น 1: PD → EX เมื่อ validity window หมดแล้ว
	res, err := db.Exec(`
		UPDATE quotations
		SET status = 'EX', log_timestamp = NOW()
		WHERE status = 'PD'
		  AND COALESCE(is_locked, false) = false
		  AND COALESCE(
		        valid_until,
		        create_time + COALESCE(validity_days, 7) * INTERVAL '1 day'
		      ) < NOW()
	`)
	if err != nil {
		log.Printf("[jobs/expiration] expireQuotations error: %v", err)
		return
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		log.Printf("[jobs/expiration] expired %d quotation(s) (PD→EX)", n)
	}

	// ขั้น 2: OP → CL สำหรับ RFQ ที่ไม่มี quotation สถานะ PD หรือ AC เหลืออยู่แล้ว
	// (ลูกค้าไม่มีใบเสนอราคาที่ active ให้ยอมรับได้ ควรปิดรับข้อเสนอเพื่อ UX ที่ชัดเจน)
	res2, err2 := db.Exec(`
		UPDATE rfqs
		SET status = 'CL', updated_at = NOW()
		WHERE status = 'OP'
		  AND NOT EXISTS (
		        SELECT 1 FROM quotations q
		        WHERE q.rfq_id = rfqs.rfq_id
		          AND q.status IN ('PD', 'AC')
		  )
	`)
	if err2 != nil {
		log.Printf("[jobs/expiration] expireRFQs (OP→CL) error: %v", err2)
		return
	}
	n2, _ := res2.RowsAffected()
	if n2 > 0 {
		log.Printf("[jobs/expiration] auto-closed %d RFQ(s) after all quotations expired (OP→CL)", n2)
	}
}

func expirePendingDeposits(db *sqlx.DB) {
	type orderRow struct {
		OrderID int64 `db:"order_id"`
	}

	// ขั้น 1: PP → PE เมื่อครบกำหนดชำระ
	var peRows []orderRow
	err := db.Select(&peRows, `
		WITH expired_orders AS (
			SELECT o.order_id
			FROM orders o
			WHERE o.status = 'PP'
			  AND COALESCE(
				(
					SELECT ps.due_date::timestamp + TIME '23:59:59'
					FROM payment_schedules ps
					WHERE ps.order_id = o.order_id
					ORDER BY ps.installment_no ASC, ps.schedule_id ASC
					LIMIT 1
				),
				o.created_at + INTERVAL '3 days'
			  ) < NOW()
		)
		UPDATE orders o
		SET status = 'PE',
		    updated_at = NOW()
		FROM expired_orders e
		WHERE o.order_id = e.order_id
		RETURNING o.order_id
	`)
	if err != nil {
		log.Printf("[jobs/expiration] expirePendingDeposits (PP→PE) error: %v", err)
		return
	}
	for _, row := range peRows {
		payload, _ := json.Marshal(map[string]interface{}{"order_id": row.OrderID})
		if _, err := db.Exec(`INSERT INTO domain_events (event_type, payload) VALUES ($1, $2)`, "order.deposit_expired", payload); err != nil {
			log.Printf("[jobs/expiration] deposit_expired event error (order %d): %v", row.OrderID, err)
		}
	}
	if len(peRows) > 0 {
		log.Printf("[jobs/expiration] expired %d pending deposit order(s) PP→PE", len(peRows))
	}

	// ขั้น 2: PE → CL เมื่อผ่าน grace period (due_date + 3 วัน) แล้วยังไม่ได้ชำระ
	// พร้อมกันนั้น unlock quotation กลับเป็น PD เพื่อให้ลูกค้า re-order ได้จาก quotation ใบเดิม
	type clOrderRow struct {
		OrderID int64 `db:"order_id"`
		QuoteID int64 `db:"quote_id"`
	}
	var clRows []clOrderRow
	err = db.Select(&clRows, `
		WITH overdue_orders AS (
			SELECT o.order_id, o.quote_id
			FROM orders o
			WHERE o.status = 'PE'
			  AND COALESCE(
				(
					SELECT ps.due_date::timestamp + TIME '23:59:59'
					FROM payment_schedules ps
					WHERE ps.order_id = o.order_id
					ORDER BY ps.installment_no ASC, ps.schedule_id ASC
					LIMIT 1
				),
				o.created_at + INTERVAL '3 days'
			  ) + INTERVAL '3 days' < NOW()
		)
		UPDATE orders o
		SET status = 'CL',
		    updated_at = NOW()
		FROM overdue_orders e
		WHERE o.order_id = e.order_id
		RETURNING o.order_id, o.quote_id
	`)
	if err != nil {
		log.Printf("[jobs/expiration] cancelExpiredDeposits (PE→CL) error: %v", err)
		return
	}
	for _, row := range clRows {
		// Unlock quotation กลับเป็น PD ให้ลูกค้า re-order ได้จาก quotation ใบเดิม
		// (quotation จะ expire เองตาม logic expireQuotations ถ้าไม่มีการสั่งซื้อใหม่ภายใน 7 วัน)
		if _, err := db.Exec(`
			UPDATE quotations
			SET status = 'PD', is_locked = FALSE, log_timestamp = NOW()
			WHERE quote_id = $1 AND status = 'AC'
		`, row.QuoteID); err != nil {
			log.Printf("[jobs/expiration] unlock quotation error (order %d, quote %d): %v", row.OrderID, row.QuoteID, err)
		}
		payload, _ := json.Marshal(map[string]interface{}{"order_id": row.OrderID, "quote_id": row.QuoteID})
		if _, err := db.Exec(`INSERT INTO domain_events (event_type, payload) VALUES ($1, $2)`, "order.auto_cancelled", payload); err != nil {
			log.Printf("[jobs/expiration] auto_cancelled event error (order %d): %v", row.OrderID, err)
		}
	}
	if len(clRows) > 0 {
		log.Printf("[jobs/expiration] auto-cancelled %d expired deposit order(s) PE→CL (quotations unlocked)", len(clRows))
	}
}

// --------------------------------------------------------------------------
// Auto-matching notification job — runs every 5 minutes
// --------------------------------------------------------------------------

// runMatchingNotifications checks for new open RFQs and sends a notification
// to each factory whose category mapping matches the RFQ.
func runMatchingNotifications(db *sqlx.DB) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sendMatchingNotifications(db)
	}
}

// sendMatchingNotifications finds RFQs created in the last 6 minutes (slightly
// longer than the ticker interval to tolerate drift) and notifies matching factories
// that have not yet received a notification for that RFQ.
func sendMatchingNotifications(db *sqlx.DB) {
	type rfqRow struct {
		RFQID         int64  `db:"rfq_id"`
		Title         string `db:"title"`
		SubCategoryID *int64 `db:"sub_category_id"`
		CategoryID    int64  `db:"category_id"`
	}

	var newRFQs []rfqRow
	err := db.Select(&newRFQs, `
		SELECT rfq_id, title, sub_category_id, category_id
		FROM rfqs
		WHERE status = 'OP'
		  AND created_at >= NOW() - INTERVAL '6 minutes'
	`)
	if err != nil {
		log.Printf("[jobs/matching] query new RFQs error: %v", err)
		return
	}
	if len(newRFQs) == 0 {
		return
	}

	for _, rfq := range newRFQs {
		notifyMatchingFactories(db, rfq.RFQID, rfq.CategoryID, rfq.SubCategoryID, rfq.Title)
	}
}

func notifyMatchingFactories(db *sqlx.DB, rfqID, categoryID int64, subCategoryID *int64, rfqTitle string) {
	// Find factories whose category (and optional sub-category) matches this RFQ
	// and have not already been notified for this RFQ.
	type factoryRow struct {
		UserID int64 `db:"user_id"`
	}
	var factories []factoryRow

	var subCatArg interface{}
	if subCategoryID != nil {
		subCatArg = *subCategoryID
	}

	err := db.Select(&factories, `
		SELECT DISTINCT mfc.factory_id AS user_id
		FROM map_factory_categories mfc
		INNER JOIN users u ON u.user_id = mfc.factory_id AND u.role = 'FT' AND u.is_active = TRUE
		WHERE mfc.category_id = $1
		  AND (
			$2::bigint IS NULL
			OR EXISTS (
				SELECT 1 FROM map_factory_sub_categories ms
				WHERE ms.factory_id = mfc.factory_id
				  AND ms.sub_category_id = $2
			)
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM notifications n
			WHERE n.user_id = mfc.factory_id
			  AND n.type = 'RFQ'
			  AND n.link_to = '/factory/rfqs/' || $3::text
		  )
	`, categoryID, subCatArg, rfqID)
	if err != nil {
		log.Printf("[jobs/matching] query matching factories error: %v", err)
		return
	}
	if len(factories) == 0 {
		return
	}

	title := "มี RFQ ใหม่ตรงกับหมวดของคุณ"
	body := "มีคำขอ RFQ ใหม่: " + rfqTitle + " กดเพื่อดูรายละเอียด"

	for _, f := range factories {
		_, err := db.Exec(`
			INSERT INTO notifications (user_id, type, title, message, link_to, is_read)
			VALUES ($1, 'RFQ', $2, $3, '/factory/rfqs/' || $4::text, FALSE)
		`, f.UserID, title, body, rfqID)
		if err != nil {
			log.Printf("[jobs/matching] insert notification error (factory %d, rfq %d): %v", f.UserID, rfqID, err)
		}
	}
	log.Printf("[jobs/matching] sent notifications for RFQ %d to %d factory/factories", rfqID, len(factories))
}
