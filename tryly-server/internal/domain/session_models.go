package domain

type SessionCurrentUser struct {
	ID          int64  `json:"id"`
	Role        string `json:"role"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	MemberSince string `json:"member_since"`
}

type SessionWallet struct {
	Balance        float64 `json:"balance"`
	PendingBalance float64 `json:"pending_balance"`
}

type SessionOffer struct {
	QuoteID         int64    `json:"quote_id"`
	FactoryID       int64    `json:"factory_id"`
	FactoryName     string   `json:"factory_name"`
	FactoryImageURL *string  `json:"factory_image_url"`
	FactoryVerified bool     `json:"factory_verified"`
	PricePerPiece   *float64 `json:"price_per_piece"`
	GrandTotal      *float64 `json:"grand_total"`
	LeadTimeDays    *int64   `json:"lead_time_days"`
	Status          string   `json:"status"`
}

type SessionRFQ struct {
	RFQID        int64          `json:"rfq_id"`
	Title        string         `json:"title"`
	Status       string         `json:"status"`
	Quantity     *int64         `json:"quantity"`
	TargetPrice  *float64       `json:"target_price"`
	CategoryName string         `json:"category_name"`
	CreatedAt    string         `json:"created_at"`
	OfferCount   int64          `json:"offer_count"`
	Offers       []SessionOffer `json:"offers"`
}

type SessionOrder struct {
	OrderID           int64    `json:"order_id"`
	Title             string   `json:"title"`
	FactoryID         int64    `json:"factory_id"`
	FactoryName       string   `json:"factory_name"`
	FactoryImageURL   *string  `json:"factory_image_url"`
	Status            string   `json:"status"`
	TotalAmount       *float64 `json:"total_amount"`
	DepositAmount     *float64 `json:"deposit_amount"`
	EstimatedDelivery *string  `json:"estimated_delivery"`
	CreatedAt         string   `json:"created_at"`
}

type SessionThread struct {
	ConvID             int64   `json:"conv_id"`
	CounterpartID      int64   `json:"counterpart_id"`
	CounterpartName    string  `json:"counterpart_name"`
	CounterpartImageURL *string `json:"counterpart_image_url"`
	LastMessage        string  `json:"last_message"`
	LastMessageAt      string  `json:"last_message_at"`
	Unread             int64   `json:"unread"`
}

type SessionResponse struct {
	CurrentUser *SessionCurrentUser `json:"currentUser"`
	Wallet      *SessionWallet      `json:"wallet"`
	RFQs        []SessionRFQ        `json:"rfqs"`
	Orders      []SessionOrder      `json:"orders"`
	Threads     []SessionThread     `json:"threads"`
	Favorites   []int64             `json:"favorites"`
}
