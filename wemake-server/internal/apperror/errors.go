package apperror

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// ============= COMMON ERRORS =============

var (
	ErrNotFound        = errors.New("not_found")
	ErrInvalidInput    = errors.New("invalid_input")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrAlreadyExists   = errors.New("already_exists")
	ErrInternalServer  = errors.New("internal_server_error")
)

// ============= AUTH ERRORS =============

var (
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrUserNotFound       = errors.New("user_not_found")
	ErrUserInactive       = errors.New("user_inactive")
	ErrEmailExists        = errors.New("email_already_exists")
	ErrInvalidRole        = errors.New("invalid_role")
	ErrMissingRoleData    = errors.New("missing_role_data")
	ErrInvalidResetToken  = errors.New("invalid_reset_token")
)

// ============= FACTORY ERRORS =============

var (
	ErrFactoryNotFound      = errors.New("factory_not_found")
	ErrFactorySuspended     = errors.New("factory_suspended")
	ErrFactoryInactive      = errors.New("factory_inactive")
	ErrInvalidFactoryType   = errors.New("invalid_factory_type")
	ErrUnauthorizedFactory  = errors.New("unauthorized_factory")
)

// ============= QUOTATION ERRORS =============

var (
	ErrQuotationNotFound        = errors.New("quotation_not_found")
	ErrQuotationLocked          = errors.New("quotation_locked")
	ErrQuotationExpired         = errors.New("quotation_expired")
	ErrInvalidShippingMethod    = errors.New("invalid_shipping_method")
	ErrInvalidPaymentTerms      = errors.New("invalid_payment_terms")
	ErrFactoryHighlightInvalid  = errors.New("factory_highlight_invalid")
	ErrInvalidLineItem          = errors.New("invalid_line_item")
)

// ============= RFQ ERRORS =============

var (
	ErrRFQNotFound              = errors.New("rfq_not_found")
	ErrInvalidCategory          = errors.New("invalid_category")
	ErrInvalidSubCategory       = errors.New("invalid_sub_category")
	ErrRFQInactive              = errors.New("rfq_inactive")
	ErrRFQClosed                = errors.New("rfq_closed")
	ErrRFQCannotCancel          = errors.New("rfq_cannot_cancel")
	ErrRFQDetailsRequired       = errors.New("rfq_details_required")
	ErrRFQDetailsTooShort       = errors.New("rfq_details_too_short")
	ErrRFQKindInvalid           = errors.New("rfq_kind_invalid")
	ErrRFQSampleQtyInvalid      = errors.New("rfq_sample_qty_invalid")
	ErrRFQWrongScope            = errors.New("rfq_wrong_scope")
	ErrMaxRFQReferenceImages    = errors.New("max_rfq_reference_images")
	ErrRFQInspectionInvalid     = errors.New("rfq_inspection_invalid")
)

// ============= ORDER ERRORS =============

var (
	ErrOrderNotFound          = errors.New("order_not_found")
	ErrOrderInvalidStatus     = errors.New("order_invalid_status")
	ErrOrderCannotCancel      = errors.New("order_cannot_cancel")
	ErrOrderCannotShip        = errors.New("order_cannot_ship")
	ErrOrderPaymentIncomplete = errors.New("order_payment_incomplete")
	ErrInvalidTrackingInfo    = errors.New("invalid_tracking_info")
	ErrOrderAlreadyShipped    = errors.New("order_already_shipped")
)

// ============= PAYMENT ERRORS =============

var (
	ErrPaymentNotFound       = errors.New("payment_not_found")
	ErrInvalidAmount         = errors.New("invalid_amount")
	ErrPaymentProcessing     = errors.New("payment_processing")
	ErrPaymentFailed         = errors.New("payment_failed")
	ErrInsufficientBalance   = errors.New("insufficient_balance")
	ErrWalletNotFound        = errors.New("wallet_not_found")
)

// ============= ADDRESS ERRORS =============

var (
	ErrAddressNotFound   = errors.New("address_not_found")
	ErrInvalidProvince   = errors.New("invalid_province")
	ErrInvalidDistrict   = errors.New("invalid_district")
	ErrInvalidSubDistrict = errors.New("invalid_sub_district")
)

// ============= SHOWCASE ERRORS =============

var (
	ErrShowcaseNotFound  = errors.New("showcase_not_found")
	ErrInvalidShowcase   = errors.New("invalid_showcase")
	ErrShowcaseInactive  = errors.New("showcase_inactive")
)

// ============= CERTIFICATE ERRORS =============

var (
	ErrCertificateNotFound = errors.New("certificate_not_found")
	ErrInvalidCertificate  = errors.New("invalid_certificate")
)

// ============= HTTP STATUS MAPPING =============

// ErrorStatusMap maps errors to HTTP status codes
var ErrorStatusMap = map[error]int{
	// Common
	ErrNotFound:       fiber.StatusNotFound,
	ErrInvalidInput:   fiber.StatusBadRequest,
	ErrUnauthorized:   fiber.StatusUnauthorized,
	ErrForbidden:      fiber.StatusForbidden,
	ErrAlreadyExists:  fiber.StatusConflict,
	ErrInternalServer: fiber.StatusInternalServerError,

	// Auth
	ErrInvalidCredentials: fiber.StatusUnauthorized,
	ErrUserNotFound:       fiber.StatusNotFound,
	ErrUserInactive:       fiber.StatusForbidden,
	ErrEmailExists:        fiber.StatusConflict,
	ErrInvalidRole:        fiber.StatusBadRequest,
	ErrMissingRoleData:    fiber.StatusBadRequest,
	ErrInvalidResetToken:  fiber.StatusBadRequest,

	// Factory
	ErrFactoryNotFound:    fiber.StatusNotFound,
	ErrFactorySuspended:   fiber.StatusForbidden,
	ErrFactoryInactive:    fiber.StatusForbidden,
	ErrInvalidFactoryType: fiber.StatusBadRequest,

	// Quotation
	ErrQuotationNotFound:      fiber.StatusNotFound,
	ErrQuotationLocked:        fiber.StatusForbidden,
	ErrQuotationExpired:       fiber.StatusBadRequest,
	ErrInvalidShippingMethod:  fiber.StatusBadRequest,
	ErrInvalidPaymentTerms:    fiber.StatusBadRequest,
	ErrFactoryHighlightInvalid: fiber.StatusBadRequest,

	// RFQ
	ErrRFQNotFound:        fiber.StatusNotFound,
	ErrInvalidCategory:    fiber.StatusBadRequest,
	ErrInvalidSubCategory: fiber.StatusBadRequest,

	// Order
	ErrOrderNotFound:      fiber.StatusNotFound,
	ErrOrderInvalidStatus: fiber.StatusBadRequest,

	// Payment
	ErrPaymentNotFound:     fiber.StatusNotFound,
	ErrInvalidAmount:       fiber.StatusBadRequest,
	ErrInsufficientBalance: fiber.StatusBadRequest,

	// Address
	ErrAddressNotFound:   fiber.StatusNotFound,
	ErrInvalidProvince:   fiber.StatusBadRequest,
	ErrInvalidDistrict:   fiber.StatusBadRequest,
	ErrInvalidSubDistrict: fiber.StatusBadRequest,

	// Showcase
	ErrShowcaseNotFound: fiber.StatusNotFound,

	// Certificate
	ErrCertificateNotFound: fiber.StatusNotFound,
}

// GetHTTPStatus ดึง HTTP status code จาก error
func GetHTTPStatus(err error) int {
	if status, exists := ErrorStatusMap[err]; exists {
		return status
	}
	return fiber.StatusInternalServerError
}

// ============= ERROR CODE MAPPING (for API responses) =============

// ErrorCodeMap maps errors to API error codes
var ErrorCodeMap = map[error]string{
	ErrNotFound:        "NOT_FOUND",
	ErrInvalidInput:    "INVALID_INPUT",
	ErrUnauthorized:    "UNAUTHORIZED",
	ErrForbidden:       "FORBIDDEN",
	ErrAlreadyExists:   "ALREADY_EXISTS",
	ErrInternalServer:  "INTERNAL_ERROR",

	ErrInvalidCredentials: "INVALID_CREDENTIALS",
	ErrUserNotFound:       "USER_NOT_FOUND",
	ErrEmailExists:        "EMAIL_ALREADY_EXISTS",

	ErrFactoryNotFound:    "FACTORY_NOT_FOUND",
	ErrFactorySuspended:   "FACTORY_SUSPENDED",

	ErrQuotationNotFound:  "QUOTATION_NOT_FOUND",
	ErrRFQNotFound:        "RFQ_NOT_FOUND",
	ErrOrderNotFound:      "ORDER_NOT_FOUND",
	ErrPaymentNotFound:    "PAYMENT_NOT_FOUND",
	ErrAddressNotFound:    "ADDRESS_NOT_FOUND",
	ErrShowcaseNotFound:   "SHOWCASE_NOT_FOUND",
	ErrCertificateNotFound: "CERTIFICATE_NOT_FOUND",
}

// GetErrorCode ดึง error code จาก error
func GetErrorCode(err error) string {
	if code, exists := ErrorCodeMap[err]; exists {
		return code
	}
	return "INTERNAL_ERROR"
}
