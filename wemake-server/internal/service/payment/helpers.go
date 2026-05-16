package payment

import (
	"encoding/json"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/domainutil"
)

var (
	ErrPaymentAmountMismatch = errors.New("payment amount does not match order amount for payment type")
	ErrDepositAlreadyPaid    = errors.New("DEPOSIT_ALREADY_PAID")
	ErrDepositExpired        = errors.New("DEPOSIT_EXPIRED")
)

func normalizeOrderStatus(status string) string {
	switch domainutil.NormalizeStatus(status) {
	case "CC":
		return "CN"
	default:
		return domainutil.NormalizeStatus(status)
	}
}

func roundCurrency(v float64) float64 {
	return domainutil.RoundMoney(v)
}

func insertDomainEventTx(tx *sqlx.Tx, eventType string, payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO domain_events (event_type, payload) VALUES ($1, $2)`, eventType, b)
	return err
}
