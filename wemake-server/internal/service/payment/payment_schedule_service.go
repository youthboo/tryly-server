package payment

import (
	"errors"
	"strings"

	"github.com/yourusername/wemake/internal/domain"
	paymentrepo "github.com/yourusername/wemake/internal/repository/payment"
)

var ErrInvalidScheduleStatus = errors.New("status must be PD or OD")

type PaymentScheduleService struct {
	repo *paymentrepo.PaymentScheduleRepository
}

func NewPaymentScheduleService(repo *paymentrepo.PaymentScheduleRepository) *PaymentScheduleService {
	return &PaymentScheduleService{repo: repo}
}

func (s *PaymentScheduleService) ListByOrderID(orderID int64) ([]domain.PaymentSchedule, error) {
	return s.repo.ListByOrderID(orderID)
}

func (s *PaymentScheduleService) CreateSchedule(sc *domain.PaymentSchedule) error {
	return s.repo.Create(sc)
}

func (s *PaymentScheduleService) PatchStatus(scheduleID int64, status string) error {
	status = strings.ToUpper(strings.TrimSpace(status))
	if status != "PD" && status != "OD" {
		return ErrInvalidScheduleStatus
	}
	return s.repo.PatchStatus(scheduleID, status)
}
