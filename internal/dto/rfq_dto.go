package dto

// RFQ Request DTOs
type CreateRFQRequest struct {
	CategoryID             int64    `json:"category_id" validate:"gt=0"`
	SubCategoryID          *int64   `json:"sub_category_id"`
	Title                  string   `json:"title" validate:"notblank"`
	Description            string   `json:"description"`
	Quantity               int64    `json:"quantity" validate:"gt=0"`
	Unit                   string   `json:"unit"`
	Details                string   `json:"details"`
	AddressID              int64    `json:"address_id" validate:"gt=0"`
	ShippingMethodID       *int64   `json:"shipping_method_id"`
	MaterialGrade          *string  `json:"material_grade"`
	TargetPrice            *float64 `json:"target_price"`
	TargetLeadTimeDays     *int     `json:"target_lead_time_days"`
	RequiredDeliveryDate   *string  `json:"required_delivery_date"`
	DeliveryAddressID      *int64   `json:"delivery_address_id"`
	CertificationsRequired []string `json:"certifications_required"`
	ReferenceImages        []string `json:"reference_images"`
	RequestKind            string   `json:"request_kind"`
}

type PatchRFQRequest struct {
	CategoryID             *int64   `json:"category_id"`
	SubCategoryID          *int64   `json:"sub_category_id"`
	Title                  *string  `json:"title"`
	Description            *string  `json:"description"`
	Quantity               *int64   `json:"quantity"`
	Unit                   *string  `json:"unit"`
	Details                *string  `json:"details"`
	MaterialGrade          *string  `json:"material_grade"`
	TargetPrice            *float64 `json:"target_price"`
	TargetLeadTimeDays     *int     `json:"target_lead_time_days"`
	RequiredDeliveryDate   *string  `json:"required_delivery_date"`
	CertificationsRequired []string `json:"certifications_required"`
	ReferenceImages        []string `json:"reference_images"`
}

type CancelRFQRequest struct {
	Reason string `json:"reason" validate:"notblank"`
}

type PreviewFactoriesRequest struct {
	CategoryID    int64   `json:"category_id" validate:"gt=0"`
	SubCategoryID *int64  `json:"sub_category_id"`
	Quantity      int64   `json:"quantity" validate:"gt=0"`
	TargetPrice   *float64 `json:"target_price"`
}
