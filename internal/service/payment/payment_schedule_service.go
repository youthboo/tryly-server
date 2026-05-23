package payment

import (
	"errors"

	"github.com/yourusername/wemake/internal/domain"
	domainstatus "github.com/yourusername/wemake/internal/domain/status"
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
	status = domainstatus.NormalizeCode(status)
	if !domainstatus.IsValidPaymentSchedulePatchStatus(status) {
		return ErrInvalidScheduleStatus
	}
	return s.repo.PatchStatus(scheduleID, status)
}
