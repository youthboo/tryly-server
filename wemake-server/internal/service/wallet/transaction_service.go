package wallet

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/wemake/internal/domain"
	walletrepo "github.com/yourusername/wemake/internal/repository/wallet"
)

type TransactionService struct {
	repo *walletrepo.TransactionRepository
}

func NewTransactionService(repo *walletrepo.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) Create(item *domain.Transaction) error {
	now := time.Now()
	item.TxID = "tx-" + uuid.NewString()
	item.Type = strings.TrimSpace(strings.ToUpper(item.Type))
	item.Status = strings.TrimSpace(strings.ToUpper(item.Status))
	item.CreatedAt = now
	item.UpdatedAt = now
	item.UploadedAt = now
	return s.repo.Create(item)
}

func (s *TransactionService) List(filters walletrepo.TransactionFilters) ([]domain.Transaction, error) {
	return s.repo.List(filters)
}

func (s *TransactionService) PatchStatus(txID string, status string) error {
	return s.repo.PatchStatus(strings.TrimSpace(txID), strings.TrimSpace(strings.ToUpper(status)))
}
