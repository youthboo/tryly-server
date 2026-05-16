package dto

// Order Request DTOs
type CreateOrderFromQuoteRequest struct {
	QuotationID int64 `json:"quote_id" validate:"gt=0"`
}

type CreateOrderRequest struct {
	QuotationID int64  `json:"quotation_id" validate:"gt=0"`
	Quantity    int64  `json:"quantity" validate:"gt=0"`
	AddressID   int64  `json:"address_id" validate:"gt=0"`
	Notes       string `json:"notes"`
}

type BulkCheckoutRequest struct {
	OrderItems []BulkCheckoutItem `json:"order_items"`
	AddressID  int64              `json:"address_id" validate:"gt=0"`
}

type BulkCheckoutItem struct {
	QuotationID int64 `json:"quotation_id" validate:"gt=0"`
	Quantity    int64 `json:"quantity" validate:"gt=0"`
}

type ShipOrderRequest struct {
	TrackingNo string `json:"tracking_no" validate:"notblank"`
	Courier    string `json:"courier" validate:"notblank"`
}

type ConfirmReceiptRequest struct {
	Note       string  `json:"note"`
	ReceivedAt *string `json:"received_at"`
}

type CancelOrderRequest struct {
	Reason string `json:"reason" validate:"notblank"`
}

type MarkShippedRequest struct {
	TrackingNo string `json:"tracking_no"`
	Courier    string `json:"courier"`
}

type CreateDisputeRequest struct {
	Category    string `json:"category" validate:"notblank"`
	Title       string `json:"title" validate:"notblank"`
	Description string `json:"description" validate:"notblank"`
	ImageURLs   []string `json:"image_urls"`
}

type PatchDisputeStatusRequest struct {
	Status   string `json:"status" validate:"notblank"`
	Comments *string `json:"comments"`
}

type VerifyPaymentRequest struct {
	ProofOfPayment string `json:"proof_of_payment" validate:"notblank"`
}
