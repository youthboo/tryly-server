package payment

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/helper"
)

var (
	ErrPaymentAmountMismatch = errors.New("payment amount does not match order amount for payment type")
	ErrDepositAlreadyPaid    = errors.New("DEPOSIT_ALREADY_PAID")
	ErrDepositExpired        = errors.New("DEPOSIT_EXPIRED")
)

func normalizeOrderStatus(status string) string {
	return helper.NormalizeOrderStatus(status)
}

func insertDomainEventTx(tx *sqlx.Tx, eventType string, payload interface{}) error {
	return helper.InsertDomainEventTx(tx, eventType, payload)
}
