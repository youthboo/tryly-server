package status

import (
	"strings"

	"github.com/yourusername/wemake/internal/domain"
)

func NormalizeCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func NormalizeOrder(value string) string {
	switch NormalizeCode(value) {
	case "CC":
		return "CN"
	default:
		return NormalizeCode(value)
	}
}

func IsValidOrder(value string) bool {
	switch NormalizeOrder(value) {
	case domain.OrderStatusPaymentPending,
		domain.OrderStatusPaymentExpired,
		domain.OrderStatusPaymentDone,
		domain.OrderStatusProduction,
		domain.OrderStatusWaitingFinalPayment,
		domain.OrderStatusQualityCheck,
		domain.OrderStatusShipping,
		domain.OrderStatusDelivered,
		domain.OrderStatusAccepted,
		domain.OrderStatusComplete,
		domain.OrderStatusCancelled:
		return true
	default:
		return false
	}
}

func OrderLabelTH(value string) string {
	switch NormalizeOrder(value) {
	case domain.OrderStatusPaymentPending:
		return "รอชำระเงิน"
	case domain.OrderStatusPaymentExpired:
		return "หมดกำหนดชำระ"
	case domain.OrderStatusPaymentDone:
		return "ชำระเงินแล้ว รอเริ่มผลิต"
	case domain.OrderStatusProduction:
		return "กำลังผลิต"
	case domain.OrderStatusQualityCheck:
		return "ตรวจสอบคุณภาพ"
	case domain.OrderStatusShipping:
		return "จัดส่งแล้ว"
	case domain.OrderStatusComplete:
		return "เสร็จสิ้น"
	case domain.OrderStatusCancelled:
		return "ยกเลิก"
	default:
		return NormalizeOrder(value)
	}
}

func IsDepositPaidOrBeyondOrder(value string) bool {
	switch NormalizeOrder(value) {
	case domain.OrderStatusPaymentDone,
		domain.OrderStatusProduction,
		domain.OrderStatusQualityCheck,
		domain.OrderStatusShipping,
		domain.OrderStatusComplete:
		return true
	default:
		return false
	}
}

func IsDepositExpiredOrder(value string) bool {
	return NormalizeOrder(value) == domain.OrderStatusPaymentExpired
}

func IsCancellableOrder(value string) bool {
	switch NormalizeOrder(value) {
	case domain.OrderStatusPaymentExpired,
		domain.OrderStatusPaymentPending,
		domain.OrderStatusProduction,
		domain.OrderStatusWaitingFinalPayment:
		return true
	default:
		return false
	}
}

func IsValidPaymentSchedulePatchStatus(value string) bool {
	switch NormalizeCode(value) {
	case domain.PaymentScheduleStatusPaid, domain.PaymentScheduleStatusOverdue:
		return true
	default:
		return false
	}
}

func IsPreProductionOrder(value string) bool {
	switch NormalizeOrder(value) {
	case "CF", domain.OrderStatusPaymentExpired, domain.OrderStatusPaymentPending:
		return true
	default:
		return false
	}
}

func IsProductionLockedOrder(value string) bool {
	switch NormalizeOrder(value) {
	case domain.OrderStatusPaymentPending,
		domain.OrderStatusPaymentExpired,
		domain.OrderStatusCancelled,
		domain.OrderStatusComplete:
		return true
	default:
		return false
	}
}

func IsProductionReadLockedOrder(value string) bool {
	switch NormalizeOrder(value) {
	case domain.OrderStatusPaymentPending,
		domain.OrderStatusPaymentExpired,
		domain.OrderStatusCancelled:
		return true
	default:
		return false
	}
}

func ProductionLockReason(value string) string {
	switch NormalizeOrder(value) {
	case domain.OrderStatusPaymentPending:
		return "PENDING_DEPOSIT"
	case domain.OrderStatusPaymentExpired:
		return "DEPOSIT_EXPIRED"
	case domain.OrderStatusCancelled:
		return "ORDER_CANCELLED"
	default:
		return "UNKNOWN"
	}
}

func FrontendRFQ(value string, offerCount int64) string {
	switch NormalizeCode(value) {
	case "CC":
		return "cancelled"
	case "CL":
		return "completed"
	case "OP":
		if offerCount > 0 {
			return "offers_received"
		}
		return "pending"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func FrontendOrder(value string) string {
	switch NormalizeOrder(value) {
	case domain.OrderStatusProduction, domain.OrderStatusQualityCheck:
		return "in_production"
	case domain.OrderStatusShipping:
		return "shipped"
	case domain.OrderStatusComplete:
		return "completed"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func FrontendQuotation(value string) string {
	switch NormalizeCode(value) {
	case domain.QuotationStatusPrepared:
		return "pending"
	case domain.QuotationStatusAccepted:
		return "accepted"
	case domain.QuotationStatusRejected:
		return "rejected"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}
