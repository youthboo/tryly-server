package wallet

import (
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/helper"
	walletrepo "github.com/yourusername/wemake/internal/repository/wallet"
)

type SettlementService struct {
	repo *walletrepo.SettlementRepository
}

func NewSettlementService(repo *walletrepo.SettlementRepository) *SettlementService {
	return &SettlementService{repo: repo}
}

func (s *SettlementService) ListByFactoryID(factoryID int64) ([]domain.Settlement, error) {
	return s.repo.ListByFactoryID(factoryID)
}

func (s *SettlementService) GetByID(settlementID, factoryID int64) (*domain.Settlement, error) {
	return s.repo.GetByID(settlementID, factoryID)
}

func (s *SettlementService) Create(factoryID int64, orderID *int64, amount float64, note *string) (*domain.Settlement, error) {
	s2 := &domain.Settlement{
		FactoryID: factoryID,
		OrderID:   orderID,
		Amount:    helper.MoneyDecimal(amount),
		Status:    "PE",
		Note:      note,
	}
	if err := s.repo.Create(s2); err != nil {
		return nil, err
	}
	return s2, nil
}

func (s *SettlementService) UpdateStatus(settlementID int64, status string) error {
	return s.repo.UpdateStatus(settlementID, status)
}
