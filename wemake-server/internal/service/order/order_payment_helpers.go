package order

import (
	"fmt"
	"time"

	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/helper"
	orderrepo "github.com/yourusername/wemake/internal/repository/order"
	walletrepo "github.com/yourusername/wemake/internal/repository/wallet"
)

func expectedPaymentAmount(order *domain.Order, paymentType string) (float64, error) {
	switch paymentType {
	case "DP":
		return helper.DecimalToFloat(order.DepositAmount), nil
	case "FP":
		return helper.DecimalToFloat(helper.SubtractMoney(order.TotalAmount, order.DepositAmount)), nil
	default:
		return 0, ErrPaymentTypeInvalid
	}
}

func (s *OrderService) depositPaidAt(orderID int64) *time.Time {
	txType := "DP"
	status := "PT"
	items, err := s.txLedger.List(walletrepo.TransactionFilters{OrderID: &orderID, Type: &txType, Status: &status})
	if err != nil || len(items) == 0 {
		return nil
	}
	paidAt := items[0].UpdatedAt.In(thailandLocation)
	return &paidAt
}

func (s *OrderService) finalPaymentPaidAt(orderID int64) *time.Time {
	txType := "FP"
	status := "PT"
	items, err := s.txLedger.List(walletrepo.TransactionFilters{OrderID: &orderID, Type: &txType, Status: &status})
	if err != nil || len(items) == 0 {
		return nil
	}
	paidAt := items[0].UpdatedAt.In(thailandLocation)
	return &paidAt
}

func deriveDepositDueDate(row *orderrepo.OrderDetailRow) *time.Time {
	if row.DepositScheduleDue != nil && !row.DepositScheduleDue.IsZero() {
		due := row.DepositScheduleDue.In(thailandLocation)
		due = time.Date(due.Year(), due.Month(), due.Day(), 23, 59, 59, 0, thailandLocation)
		return &due
	}
	due := deriveDefaultDepositDueTimestamp(row.CreatedAt)
	return &due
}

func buildNextAction(row *orderrepo.OrderDetailRow, status string, depositDueDate, depositPaidAt, finalPaidAt *time.Time, nowTH time.Time) *domain.OrderNextAction {
	switch status {
	case "PP":
		return &domain.OrderNextAction{
			Actor:      "CUSTOMER",
			Type:       "PAY_FULL_AMOUNT",
			Amount:     helper.MoneyDecimal(row.TotalAmount),
			Currency:   "THB",
			DueDate:    depositDueDate,
			CTAURL:     fmt.Sprintf("/orders/%d/payment?stage=full", row.OrderID),
			CTALabelTH: "ชำระเงินเต็มจำนวน",
		}
	case "PE", "PR", "QC", "SH", "CP", "CN", "CC":
		return nil
	}
	return nil
}

func buildPaymentSchedule(row *orderrepo.OrderDetailRow, status string, depositDueDate, depositPaidAt, finalPaidAt *time.Time) []domain.OrderPaymentScheduleItem {
	total := row.TotalAmount
	paidStatus := "PENDING"
	if depositPaidAt != nil || finalPaidAt != nil ||
		status == "PD" || status == "PR" || status == "QC" || status == "SH" || status == "CP" {
		paidStatus = "PAID"
	} else if status == "PE" {
		paidStatus = "OVERDUE"
	}

	return []domain.OrderPaymentScheduleItem{
		{
			Stage:   domain.PaymentStageFullPayment,
			Percent: helper.MoneyDecimal(100),
			Amount:  helper.MoneyDecimal(total),
			Status:  paidStatus,
			DueDate: depositDueDate,
			PaidAt:  depositPaidAt,
		},
	}
}

func orderStatusLabelTH(status string) string {
	switch status {
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
		return status
	}
}

func timePtrInTH(v *time.Time) *time.Time {
	if v == nil {
		return nil
	}
	t := v.In(thailandLocation)
	return &t
}

func (s *OrderService) ensureDepositPayable(order *domain.Order) error {
	status := helper.NormalizeOrderStatus(order.Status)
	if status == "PD" || status == "PR" || status == "QC" || status == "SH" || status == "CP" {
		return ErrDepositAlreadyPaid
	}
	if status == "PE" {
		return ErrDepositExpired
	}
	dueDate := s.lookupDepositDueDate(order)
	if dueDate != nil && time.Now().In(thailandLocation).After(*dueDate) {
		return ErrDepositExpired
	}
	return nil
}

func (s *OrderService) lookupDepositDueDate(order *domain.Order) *time.Time {
	if s.schedules != nil {
		items, err := s.schedules.ListByOrderID(order.OrderID)
		if err == nil {
			for _, item := range items {
				if item.InstallmentNo == 1 {
					due := item.DueDate.In(thailandLocation)
					due = time.Date(due.Year(), due.Month(), due.Day(), 23, 59, 59, 0, thailandLocation)
					return &due
				}
			}
		}
	}
	detailRow := &orderrepo.OrderDetailRow{
		CreatedAt:          order.CreatedAt,
		DepositScheduleDue: nil,
	}
	return deriveDepositDueDate(detailRow)
}

func deriveDefaultDepositScheduleDate(createdAt time.Time) time.Time {
	due := createdAt.In(thailandLocation).AddDate(0, 0, 3)
	return time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, thailandLocation)
}

func deriveDefaultDepositDueTimestamp(createdAt time.Time) time.Time {
	due := deriveDefaultDepositScheduleDate(createdAt)
	return time.Date(due.Year(), due.Month(), due.Day(), 23, 59, 59, 0, thailandLocation)
}
