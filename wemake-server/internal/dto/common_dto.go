package dto

// Common Query Parameters DTOs
type PaginationParams struct {
	Page  int `query:"page" validate:"gt=0"`
	Limit int `query:"limit" validate:"gt=0,lte=100"`
}

type DateRangeParams struct {
	StartDate *string `query:"start_date"`  // YYYY-MM-DD
	EndDate   *string `query:"end_date"`    // YYYY-MM-DD
}

type SortParams struct {
	SortBy    string `query:"sort_by"`     // field name to sort by
	SortOrder string `query:"sort_order"`  // "asc" or "desc"
}

// Common Response Wrappers
type ListResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page,omitempty"`
	Limit      int         `json:"limit,omitempty"`
	HasMore    bool        `json:"has_more,omitempty"`
}

type PaginatedListResponse struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

type ErrorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Upload/Media DTOs
type FileUploadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

type AvatarUploadRequest struct {
	// File is handled by FormFile
	// Fields for validation can be added here
}

// Filter DTOs (common for list endpoints)
type FactoryFilterParams struct {
	Status      *string `query:"status"`
	ProvinceID  *int64  `query:"province_id"`
	SearchQuery *string `query:"q"`
}

type RFQFilterParams struct {
	Status      *string `query:"status"`
	CategoryID  *int64  `query:"category_id"`
	SearchQuery *string `query:"q"`
}

type OrderFilterParams struct {
	Status       *string `query:"status"`
	PaymentStatus *string `query:"payment_status"`
	SearchQuery  *string `query:"q"`
}

type ShowcaseFilterParams struct {
	Status      *string `query:"status"`
	CategoryID  *int64  `query:"category_id"`
	SearchQuery *string `query:"q"`
	MinPrice    *float64 `query:"min_price"`
	MaxPrice    *float64 `query:"max_price"`
}

// Notification DTOs
type NotificationPreferencesResponse struct {
	EmailNotifications bool `json:"email_notifications"`
	SMSNotifications   bool `json:"sms_notifications"`
	PushNotifications  bool `json:"push_notifications"`
	OrderUpdates       bool `json:"order_updates"`
	PromotionalEmails  bool `json:"promotional_emails"`
	NewsletterEmails   bool `json:"newsletter_emails"`
}

// Summary DTOs
type DashboardSummary struct {
	TotalOrders       int     `json:"total_orders"`
	TotalRevenue      float64 `json:"total_revenue"`
	TotalCustomers    int     `json:"total_customers"`
	ActiveListings    int     `json:"active_listings"`
	PendingQuotations int     `json:"pending_quotations"`
}
