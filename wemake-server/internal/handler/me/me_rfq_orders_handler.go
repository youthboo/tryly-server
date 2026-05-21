package me

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/helper"
)

// RFQOrderSummary is the response shape for a single item in GET /v1/me/rfq-orders.
type RFQOrderSummary struct {
	RFQID             int64    `db:"rfq_id"             json:"rfq_id"`
	Title             string   `db:"title"              json:"title"`
	RequestKind       string   `db:"request_kind"       json:"request_kind"`
	Status            string   `db:"status"             json:"status"`
	CreatedAt         string   `db:"created_at"         json:"created_at"`
	CategoryName      *string  `db:"category_name"      json:"category_name"`
	QuotationCount    int      `db:"quotation_count"    json:"quotation_count"`
	TargetPrice       *float64 `db:"target_price"       json:"target_price,omitempty"`
	OrderID           *int64   `db:"order_id"           json:"order_id,omitempty"`
	OrderStatus       *string  `db:"order_status"       json:"order_status,omitempty"`
	TotalAmount       *float64 `db:"total_amount"       json:"total_amount,omitempty"`
	FactoryID         *int64   `db:"factory_id"         json:"factory_id,omitempty"`
	FactoryName       *string  `db:"factory_name"       json:"factory_name,omitempty"`
	EstimatedDelivery *string  `db:"estimated_delivery" json:"estimated_delivery,omitempty"`
	OrderCreatedAt    *string  `db:"order_created_at"   json:"order_created_at,omitempty"`
}

// QuotationSummary is used inside the detail response.
type QuotationSummary struct {
	QuoteID     int64    `db:"quote_id"      json:"quote_id"`
	FactoryName string   `db:"factory_name"  json:"factory_name"`
	GrandTotal  float64  `db:"grand_total"   json:"grand_total"`
	Status      string   `db:"status"        json:"status"`
	CreatedAt   string   `db:"create_time"   json:"created_at"`
}

// RFQOrderDetail is the response shape for GET /v1/me/rfq-orders/:rfq_id.
type RFQOrderDetail struct {
	RFQID              int64              `json:"rfq_id"`
	Title              string             `json:"title"`
	RequestKind        string             `json:"request_kind"`
	Status             string             `json:"status"`
	CreatedAt          string             `json:"created_at"`
	CategoryName       *string            `json:"category_name"`
	Quantity           int64              `json:"quantity"`
	Details            *string            `json:"details,omitempty"`
	TargetPrice        *float64           `json:"target_price,omitempty"`
	TargetLeadTimeDays *int               `json:"target_lead_time_days,omitempty"`
	Quotations         []QuotationSummary `json:"quotations"`
	Order              *OrderSummary      `json:"order,omitempty"`
}

// OrderSummary is the order section inside detail response.
type OrderSummary struct {
	OrderID     int64    `json:"order_id"`
	OrderStatus string   `json:"order_status"`
	TotalAmount float64  `json:"total_amount"`
	CreatedAt   string   `json:"created_at"`
}

// MeRFQOrdersHandler handles the unified /v1/me/rfq-orders endpoints.
type MeRFQOrdersHandler struct {
	db *sqlx.DB
}

// NewMeRFQOrdersHandler creates a new handler backed by a direct DB connection.
func NewMeRFQOrdersHandler(db *sqlx.DB) *MeRFQOrdersHandler {
	return &MeRFQOrdersHandler{db: db}
}

// ListRFQOrders handles GET /v1/me/rfq-orders
func (h *MeRFQOrdersHandler) ListRFQOrders(c *fiber.Ctx) error {
	userID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}

	const query = `
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
	`

	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return helper.JSONError(c, fiber.StatusInternalServerError, "failed to list rfq-orders")
	}
	defer rows.Close()

	result := make([]RFQOrderSummary, 0)
	for rows.Next() {
		var item RFQOrderSummary
		if err := rows.StructScan(&item); err != nil {
			return helper.JSONError(c, fiber.StatusInternalServerError, "failed to scan rfq-orders row")
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return helper.JSONError(c, fiber.StatusInternalServerError, "row iteration error")
	}

	return c.JSON(result)
}

// GetRFQOrderDetail handles GET /v1/me/rfq-orders/:rfq_id
func (h *MeRFQOrdersHandler) GetRFQOrderDetail(c *fiber.Ctx) error {
	userID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}
	rfqID, err := helper.RequireInt64Param(c, "rfq_id")
	if err != nil {
		return err
	}

	// Fetch base RFQ row
	type rfqRow struct {
		RFQID              int64    `db:"rfq_id"`
		Title              string   `db:"title"`
		RequestKind        string   `db:"request_kind"`
		Status             string   `db:"status"`
		CreatedAt          string   `db:"created_at"`
		CategoryName       *string  `db:"category_name"`
		Quantity           int64    `db:"quantity"`
		Details            *string  `db:"details"`
		TargetPrice        *float64 `db:"target_price"`
		TargetLeadTimeDays *int     `db:"target_lead_time_days"`
	}

	const rfqQuery = `
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
	`

	var rfq rfqRow
	if err := h.db.QueryRowx(rfqQuery, rfqID, userID).StructScan(&rfq); err != nil {
		if err == sql.ErrNoRows {
			return helper.JSONError(c, fiber.StatusNotFound, "rfq not found")
		}
		return helper.JSONError(c, fiber.StatusInternalServerError, "failed to fetch rfq")
	}

	// Fetch quotations
	const quotesQuery = `
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
	`

	quotRows, err := h.db.Queryx(quotesQuery, rfqID)
	if err != nil {
		return helper.JSONError(c, fiber.StatusInternalServerError, "failed to fetch quotations")
	}
	defer quotRows.Close()

	quotations := make([]QuotationSummary, 0)
	for quotRows.Next() {
		var qs QuotationSummary
		if err := quotRows.StructScan(&qs); err != nil {
			return helper.JSONError(c, fiber.StatusInternalServerError, "failed to scan quotation row")
		}
		quotations = append(quotations, qs)
	}
	if err := quotRows.Err(); err != nil {
		return helper.JSONError(c, fiber.StatusInternalServerError, "quotation row iteration error")
	}

	// Fetch latest order for this rfq (via quotation join)
	const orderQuery = `
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
	`

	type orderRow struct {
		OrderID     int64   `db:"order_id"`
		OrderStatus string  `db:"order_status"`
		TotalAmount float64 `db:"total_amount"`
		CreatedAt   string  `db:"created_at"`
	}

	var ord orderRow
	var orderSummary *OrderSummary
	if scanErr := h.db.QueryRowx(orderQuery, rfqID, userID).StructScan(&ord); scanErr == nil {
		orderSummary = &OrderSummary{
			OrderID:     ord.OrderID,
			OrderStatus: ord.OrderStatus,
			TotalAmount: ord.TotalAmount,
			CreatedAt:   ord.CreatedAt,
		}
	}

	detail := RFQOrderDetail{
		RFQID:              rfq.RFQID,
		Title:              rfq.Title,
		RequestKind:        rfq.RequestKind,
		Status:             rfq.Status,
		CreatedAt:          rfq.CreatedAt,
		CategoryName:       rfq.CategoryName,
		Quantity:           rfq.Quantity,
		Details:            rfq.Details,
		TargetPrice:        rfq.TargetPrice,
		TargetLeadTimeDays: rfq.TargetLeadTimeDays,
		Quotations:         quotations,
		Order:              orderSummary,
	}

	return c.JSON(detail)
}
