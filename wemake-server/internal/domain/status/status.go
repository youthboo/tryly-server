package status

import "strings"

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
	case "PP", "PE", "PD", "PR", "WF", "QC", "SH", "DL", "AC", "CP", "CN":
		return true
	default:
		return false
	}
}

func OrderLabelTH(value string) string {
	switch NormalizeOrder(value) {
	case "PP":
		return "รอชำระเงิน"
	case "PE":
		return "หมดกำหนดชำระ"
	case "PD":
		return "ชำระเงินแล้ว รอเริ่มผลิต"
	case "PR":
		return "กำลังผลิต"
	case "QC":
		return "ตรวจสอบคุณภาพ"
	case "SH":
		return "จัดส่งแล้ว"
	case "CP":
		return "เสร็จสิ้น"
	case "CN":
		return "ยกเลิก"
	default:
		return NormalizeOrder(value)
	}
}

func IsDepositPaidOrBeyondOrder(value string) bool {
	switch NormalizeOrder(value) {
	case "PD", "PR", "QC", "SH", "CP":
		return true
	default:
		return false
	}
}

func IsDepositExpiredOrder(value string) bool {
	return NormalizeOrder(value) == "PE"
}

func IsPreProductionOrder(value string) bool {
	switch NormalizeOrder(value) {
	case "CF", "PE", "PP":
		return true
	default:
		return false
	}
}

func IsProductionLockedOrder(value string) bool {
	switch NormalizeOrder(value) {
	case "PP", "PE", "CN", "CP":
		return true
	default:
		return false
	}
}

func IsProductionReadLockedOrder(value string) bool {
	switch NormalizeOrder(value) {
	case "PP", "PE", "CN":
		return true
	default:
		return false
	}
}

func ProductionLockReason(value string) string {
	switch NormalizeOrder(value) {
	case "PP":
		return "PENDING_DEPOSIT"
	case "PE":
		return "DEPOSIT_EXPIRED"
	case "CN":
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
	case "PR", "QC":
		return "in_production"
	case "SH":
		return "shipped"
	case "CP":
		return "completed"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func FrontendQuotation(value string) string {
	switch NormalizeCode(value) {
	case "PD":
		return "pending"
	case "AC":
		return "accepted"
	case "RJ":
		return "rejected"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}
