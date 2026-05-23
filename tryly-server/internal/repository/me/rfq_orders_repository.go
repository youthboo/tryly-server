package me

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type RFQOrdersRepository struct {
	db *sqlx.DB
}

func NewRFQOrdersRepository(db *sqlx.DB) *RFQOrdersRepository {
	return &RFQOrdersRepository{db: db}
}

type RFQOrderSummaryRow struct {
	RFQID             int64           `db:"rfq_id"`
	Title             string          `db:"title"`
	RequestKind       string          `db:"request_kind"`
	Status            string          `db:"status"`
	CreatedAt         string          `db:"created_at"`
	CategoryName      sql.NullString  `db:"category_name"`
	QuotationCount    int             `db:"quotation_count"`
	TargetPrice       sql.NullFloat64 `db:"target_price"`
	OrderID           sql.NullInt64   `db:"order_id"`
	OrderStatus       sql.NullString  `db:"order_status"`
	TotalAmount       sql.NullFloat64 `db:"total_amount"`
	FactoryID         sql.NullInt64   `db:"factory_id"`
	FactoryName       sql.NullString  `db:"factory_name"`
	EstimatedDelivery sql.NullString  `db:"estimated_delivery"`
	OrderCreatedAt    sql.NullString  `db:"order_created_at"`
}

type RFQOrderDetailRFQRow struct {
	RFQID              int64           `db:"rfq_id"`
	Title              string          `db:"title"`
	RequestKind        string          `db:"request_kind"`
	Status             string          `db:"status"`
	CreatedAt          string          `db:"created_at"`
	CategoryName       sql.NullString  `db:"category_name"`
	Quantity           int64           `db:"quantity"`
	Details            sql.NullString  `db:"details"`
	TargetPrice        sql.NullFloat64 `db:"target_price"`
	TargetLeadTimeDays sql.NullInt64   `db:"target_lead_time_days"`
}

type RFQOrderQuotationRow struct {
	QuoteID     int64   `db:"quote_id"`
	FactoryName string  `db:"factory_name"`
	GrandTotal  float64 `db:"grand_total"`
	Status      string  `db:"status"`
	CreatedAt   string  `db:"create_time"`
}

type RFQOrderOrderRow struct {
	OrderID     int64   `db:"order_id"`
	OrderStatus string  `db:"order_status"`
	TotalAmount float64 `db:"total_amount"`
	CreatedAt   string  `db:"created_at"`
}

func (r *RFQOrdersRepository) ListSummaries(userID int64) ([]RFQOrderSummaryRow, error) {
	var rows []RFQOrderSummaryRow
	err := r.db.Select(&rows, `
		SELECT
			r.rfq_id,
			r.title,
			COALESCE(r.request_kind, 'PR') AS request_kind,
			r.status,
			TO_CHAR(r.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at,
			lc.name AS category_name,
			COUNT(DISTINCT q.quote_id)::int AS quotation_count,
			r.target_price,
			(SELECT o2.order_id FROM orders o2 WHERE o2.quote_id IN (
				SELECT q2.quote_id FROM quotations q2 WHERE q2.rfq_id = r.rfq_id
			) AND o2.customer_id = r.user_id ORDER BY o2.created_at DESC LIMIT 1) AS order_id,
			(SELECT o2.status FROM orders o2 WHERE o2.quote_id IN (
				SELECT q2.quote_id FROM quotations q2 WHERE q2.rfq_id = r.rfq_id
			) AND o2.customer_id = r.user_id ORDER BY o2.created_at DESC LIMIT 1) AS order_status,
			(SELECT o2.total_amount FROM orders o2 WHERE o2.quote_id IN (
				SELECT q2.quote_id FROM quotations q2 WHERE q2.rfq_id = r.rfq_id
			) AND o2.customer_id = r.user_id ORDER BY o2.created_at DESC LIMIT 1) AS total_amount,
			(SELECT o2.factory_id FROM orders o2 WHERE o2.quote_id IN (
				SELECT q2.quote_id FROM quotations q2 WHERE q2.rfq_id = r.rfq_id
			) AND o2.customer_id = r.user_id ORDER BY o2.created_at DESC LIMIT 1) AS factory_id,
			(SELECT fp2.factory_name FROM orders o2
				JOIN factory_profiles fp2 ON fp2.user_id = o2.factory_id
				WHERE o2.quote_id IN (
					SELECT q2.quote_id FROM quotations q2 WHERE q2.rfq_id = r.rfq_id
				) AND o2.customer_id = r.user_id ORDER BY o2.created_at DESC LIMIT 1) AS factory_name,
			(SELECT TO_CHAR(o2.estimated_delivery, 'YYYY-MM-DD') FROM orders o2 WHERE o2.quote_id IN (
				SELECT q2.quote_id FROM quotations q2 WHERE q2.rfq_id = r.rfq_id
			) AND o2.customer_id = r.user_id ORDER BY o2.created_at DESC LIMIT 1) AS estimated_delivery,
			(SELECT TO_CHAR(o2.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') FROM orders o2 WHERE o2.quote_id IN (
				SELECT q2.quote_id FROM quotations q2 WHERE q2.rfq_id = r.rfq_id
			) AND o2.customer_id = r.user_id ORDER BY o2.created_at DESC LIMIT 1) AS order_created_at
		FROM rfqs r
		LEFT JOIN lbi_categories lc ON lc.category_id = r.category_id
		LEFT JOIN quotations q ON q.rfq_id = r.rfq_id
		LEFT JOIN orders o ON o.quote_id = q.quote_id AND o.customer_id = r.user_id
		WHERE r.user_id = $1
		GROUP BY r.rfq_id, r.title, r.request_kind, r.status, r.created_at, lc.name, r.target_price
		ORDER BY r.created_at DESC
	`, userID)
	return rows, err
}

func (r *RFQOrdersRepository) GetDetailRFQ(userID, rfqID int64) (*RFQOrderDetailRFQRow, error) {
	var row RFQOrderDetailRFQRow
	err := r.db.Get(&row, `
		SELECT
			r.rfq_id,
			r.title,
			COALESCE(r.request_kind, 'PR') AS request_kind,
			r.status,
			TO_CHAR(r.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at,
			lc.name AS category_name,
			r.quantity,
			r.details,
			r.target_price,
			r.target_lead_time_days
		FROM rfqs r
		LEFT JOIN lbi_categories lc ON lc.category_id = r.category_id
		WHERE r.rfq_id = $1 AND r.user_id = $2
	`, rfqID, userID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *RFQOrdersRepository) ListQuotations(rfqID int64) ([]RFQOrderQuotationRow, error) {
	var rows []RFQOrderQuotationRow
	err := r.db.Select(&rows, `
		SELECT
			q.quote_id,
			COALESCE(fp.factory_name, 'โรงงาน #' || q.factory_id) AS factory_name,
			q.grand_total,
			q.status,
			TO_CHAR(q.create_time, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS create_time
		FROM quotations q
		LEFT JOIN factory_profiles fp ON fp.user_id = q.factory_id
		WHERE q.rfq_id = $1
		ORDER BY q.create_time DESC
	`, rfqID)
	return rows, err
}

func (r *RFQOrdersRepository) GetLatestOrder(userID, rfqID int64) (*RFQOrderOrderRow, error) {
	var row RFQOrderOrderRow
	err := r.db.Get(&row, `
		SELECT
			o.order_id,
			o.status AS order_status,
			o.total_amount,
			TO_CHAR(o.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at
		FROM orders o
		JOIN quotations q ON q.quote_id = o.quote_id
		WHERE q.rfq_id = $1 AND o.customer_id = $2
		ORDER BY o.created_at DESC
		LIMIT 1
	`, rfqID, userID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}
