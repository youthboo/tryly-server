package wallet

import (
	"errors"
	"strings"

	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/helper"
	walletrepo "github.com/yourusername/wemake/internal/repository/wallet"
)

var ErrInsufficientFunds = errors.New("insufficient wallet funds for withdrawal")
var ErrInvalidWithdrawalStatus = errors.New("status must be AP, RJ, or CP")

type WithdrawalService struct {
	repo       *walletrepo.WithdrawalRepository
	walletRepo *walletrepo.WalletRepository
}

func NewWithdrawalService(repo *walletrepo.WithdrawalRepository, walletRepo *walletrepo.WalletRepository) *WithdrawalService {
	return &WithdrawalService{repo: repo, walletRepo: walletRepo}
}

func (s *WithdrawalService) Create(factoryID int64, amount float64, bankAccountNo, bankName, accountName string) (*domain.WithdrawalRequest, error) {
	walletID, err := s.walletRepo.GetWalletIDByUserID(factoryID)
	if err != nil {
		return nil, err
	}
	wallet, err := s.walletRepo.GetByUserID(factoryID)
	if err != nil {
		return nil, err
	}
	decimalAmount := helper.MoneyDecimal(amount)
	if helper.IsMoneyLess(wallet.GoodFund, decimalAmount) {
		return nil, ErrInsufficientFunds
	}
	w := &domain.WithdrawalRequest{
		WalletID:      *walletID,
		FactoryID:     factoryID,
		Amount:        decimalAmount,
		BankAccountNo: bankAccountNo,
		BankName:      bankName,
		AccountName:   accountName,
	}
	if err := s.repo.Create(w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *WithdrawalService) ListByFactoryID(factoryID int64) ([]domain.WithdrawalRequest, error) {
	return s.repo.ListByFactoryID(factoryID)
}

func (s *WithdrawalService) UpdateStatus(requestID int64, status string, note *string) error {
	status = strings.ToUpper(strings.TrimSpace(status))
	if status != "AP" && status != "RJ" && status != "CP" {
		return ErrInvalidWithdrawalStatus
	}
	return s.repo.UpdateStatus(requestID, status, note)
}
