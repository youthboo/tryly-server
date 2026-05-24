package wallet

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/wemake/internal/domain"
	walletrepo "github.com/yourusername/wemake/internal/repository/wallet"
	"github.com/yourusername/wemake/internal/domainutil"
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
	item.Type = domainutil.NormalizeStatus(item.Type)
	item.Status = domainutil.NormalizeStatus(item.Status)
	item.CreatedAt = now
	item.UpdatedAt = now
	item.UploadedAt = now
	return s.repo.Create(item)
}

func (s *TransactionService) List(filters walletrepo.TransactionFilters) ([]domain.Transaction, error) {
	return s.repo.List(filters)
}

func (s *TransactionService) PatchStatus(txID string, status string) error {
	return s.repo.PatchStatus(strings.TrimSpace(txID), domainutil.NormalizeStatus(status))
}
