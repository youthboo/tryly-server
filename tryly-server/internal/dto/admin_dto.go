package dto

// Admin User Management DTOs
type CreateAdminUserRequest struct {
	Email     string `json:"email" validate:"notblank"`
	Password  string `json:"password" validate:"notblank"`
	FirstName string `json:"first_name" validate:"notblank"`
	LastName  string `json:"last_name" validate:"notblank"`
	Role      string `json:"role" validate:"notblank"`
}

// Platform Config DTOs
type CreateConfigVersionRequest struct {
	DefaultCommissionRate float64  `json:"default_commission_rate"`
	PromoCommissionRate   *float64 `json:"promo_commission_rate"`
	PromoStartAt          *string  `json:"promo_start_at"`
	PromoEndAt            *string  `json:"promo_end_at"`
	PromoLabel            *string  `json:"promo_label"`
	VatRate               float64  `json:"vat_rate"`
	CurrencyCode          string   `json:"currency_code"`
}

type CreatePlatformConfigRequest struct {
	ConfigKey   string      `json:"config_key" validate:"notblank"`
	ConfigValue interface{} `json:"config_value"`
	Description *string     `json:"description"`
	ValidFrom   *string     `json:"valid_from"` // RFC3339
	ValidTo     *string     `json:"valid_to"`   // RFC3339
}

type UpdatePlatformConfigRequest struct {
	ConfigValue interface{} `json:"config_value"`
	Description *string     `json:"description"`
	ValidTo     *string     `json:"valid_to"`
}

// Commission Rules DTOs
type CreateCommissionRuleRequest struct {
	Name            string  `json:"name" validate:"notblank"`
	Description     *string `json:"description"`
	CommissionRate  float64 `json:"commission_rate" validate:"gte=0,lte=100"`
	MinOrderAmount  *float64 `json:"min_order_amount"`
	MaxOrderAmount  *float64 `json:"max_order_amount"`
	AppliesTo       string  `json:"applies_to"` // "all", "factory", "customer", etc.
	IsActive        *bool   `json:"is_active"`
}

// Commission Exemptions DTOs
type CreateCommissionExemptionRequest struct {
	UserID      int64   `json:"user_id" validate:"gt=0"`
	ExemptionID int64   `json:"exemption_id" validate:"gt=0"`
	Reason      string  `json:"reason" validate:"notblank"`
	ExemptFrom  *string `json:"exempt_from"` // RFC3339
	ExemptTo    *string `json:"exempt_to"`   // RFC3339
}

// Factory Approval DTOs
type ApproveFactoryRequest struct {
	Notes *string `json:"notes"`
}

type RejectFactoryRequest struct {
	Reason string `json:"reason" validate:"notblank"`
}

type SuspendFactoryRequest struct {
	Reason string `json:"reason" validate:"notblank"`
}

type UnsuspendFactoryRequest struct {
	Notes *string `json:"notes"`
}

type PatchFactoryVerificationRequest struct {
	Status            *string `json:"status"`
	TaxIDVerified     *bool   `json:"tax_id_verified"`
	AddressVerified   *bool   `json:"address_verified"`
	OwnershipVerified *bool   `json:"ownership_verified"`
	Notes             *string `json:"notes"`
}

// RFQ Management DTOs
type PatchRFQStatusRequest struct {
	Status string `json:"status" validate:"notblank"`
	Notes  *string `json:"notes"`
}

// Order Management DTOs
type PatchOrderStatusRequest struct {
	Status string `json:"status" validate:"notblank"`
	Notes  *string `json:"notes"`
}

// Quotation Template DTOs
type CreateQuotationTemplateRequest struct {
	Name             string                 `json:"name" validate:"notblank"`
	Description      *string                `json:"description"`
	TemplateContent  map[string]interface{} `json:"template_content"`
	IsActive         *bool                  `json:"is_active"`
}

type PatchQuotationTemplateRequest struct {
	Name             *string                `json:"name"`
	Description      *string                `json:"description"`
	TemplateContent  map[string]interface{} `json:"template_content"`
	IsActive         *bool                  `json:"is_active"`
}

// BOQ (Bill of Quantities) DTOs
type BOQPayloadRequest struct {
	Items          []interface{} `json:"items"`
	Currency       string        `json:"currency"`
	DiscountAmount float64       `json:"discount_amount"`
	VatPercent     float64       `json:"vat_percent"`
	MOQ            *int          `json:"moq"`
	LeadTimeDays   *int          `json:"lead_time_days"`
	PaymentTerms   *string       `json:"payment_terms"`
	ValidityDays   *int          `json:"validity_days"`
	Note           *string       `json:"note"`
}

type DeclineBOQRequest struct {
	Reason *string `json:"reason"`
}

type CreateBOQRequest struct {
	RFQID int64                    `json:"rfq_id" validate:"gt=0"`
	Items []BOQItem                `json:"items"`
	Notes *string                  `json:"notes"`
}

type BOQItem struct {
	Description    string   `json:"description"`
	Quantity       int64    `json:"quantity"`
	Unit           string   `json:"unit"`
	UnitPrice      float64  `json:"unit_price"`
	ItemCode       *string  `json:"item_code"`
	Specifications *string  `json:"specifications"`
}

type UpdateBOQRequest struct {
	Items []BOQItem `json:"items"`
	Notes *string   `json:"notes"`
}

type AcceptBOQRequest struct {
	Notes *string `json:"notes"`
}
