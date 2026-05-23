package domain

// MeRFQOrderSummary is one row in GET /api/v1/me/rfq-orders.
type MeRFQOrderSummary struct {
	RFQID             int64    `json:"rfq_id"`
	Title             string   `json:"title"`
	RequestKind       string   `json:"request_kind"`
	Status            string   `json:"status"`
	CreatedAt         string   `json:"created_at"`
	CategoryName      *string  `json:"category_name"`
	QuotationCount    int      `json:"quotation_count"`
	TargetPrice       *float64 `json:"target_price,omitempty"`
	OrderID           *int64   `json:"order_id,omitempty"`
	OrderStatus       *string  `json:"order_status,omitempty"`
	TotalAmount       *float64 `json:"total_amount,omitempty"`
	FactoryID         *int64   `json:"factory_id,omitempty"`
	FactoryName       *string  `json:"factory_name,omitempty"`
	EstimatedDelivery *string  `json:"estimated_delivery,omitempty"`
	OrderCreatedAt    *string  `json:"order_created_at,omitempty"`
}

// MeRFQOrderQuotation is a quotation inside GET /api/v1/me/rfq-orders/:rfq_id.
type MeRFQOrderQuotation struct {
	QuoteID     int64   `json:"quote_id"`
	FactoryName string  `json:"factory_name"`
	GrandTotal  float64 `json:"grand_total"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

// MeRFQOrderOrder is the optional order block in RFQ order detail.
type MeRFQOrderOrder struct {
	OrderID     int64   `json:"order_id"`
	OrderStatus string  `json:"order_status"`
	TotalAmount float64 `json:"total_amount"`
	CreatedAt   string  `json:"created_at"`
}

// MeRFQOrderDetail is GET /api/v1/me/rfq-orders/:rfq_id.
type MeRFQOrderDetail struct {
	RFQID              int64                   `json:"rfq_id"`
	Title              string                  `json:"title"`
	RequestKind        string                  `json:"request_kind"`
	Status             string                  `json:"status"`
	CreatedAt          string                  `json:"created_at"`
	CategoryName       *string                 `json:"category_name"`
	Quantity           int64                   `json:"quantity"`
	Details            *string                 `json:"details,omitempty"`
	TargetPrice        *float64                `json:"target_price,omitempty"`
	TargetLeadTimeDays *int                    `json:"target_lead_time_days,omitempty"`
	Quotations         []MeRFQOrderQuotation   `json:"quotations"`
	Order              *MeRFQOrderOrder        `json:"order,omitempty"`
}

// FactoryRFQBoardResponse is GET /factory/rfq-board.
type FactoryRFQBoardResponse struct {
	RFQs               []RFQ   `json:"rfqs"`
	FactoryCategoryIDs []int64 `json:"factory_category_ids"`
}

// ProfileInitResponse is GET /factories/me/profile-init.
type ProfileInitResponse struct {
	Factory          *FactoryPublicDetail    `json:"factory"`
	FactoryTypes     []LBIFactoryType      `json:"factory_types"`
	LBICategories    []Category            `json:"lbi_categories"`
	Addresses        []Address             `json:"addresses"`
	CertificateTypes []LBIMasterCertificate `json:"certificate_types"`
	SubCategories    []SubCategory         `json:"sub_categories"`
}

// RFQDetailBundle is GET /rfqs/:rfq_id/detail.
type RFQDetailBundle struct {
	RFQ             *RFQ                                `json:"rfq"`
	Quotations      []Quotation                         `json:"quotations"`
	QuoteHistories  map[string][]QuotationHistoryEntry  `json:"quote_histories"`
}
