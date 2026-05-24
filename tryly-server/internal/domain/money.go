package domain

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

// MoneyString is a response-only money value that always marshals with two
// decimal places. Keep DB/domain calculation fields as decimal.Decimal and use
// this type at response boundaries that need stable FE display formatting.
type MoneyString decimal.Decimal

func NewMoneyString(value decimal.Decimal) MoneyString {
	return MoneyString(value.Round(2))
}

func (m MoneyString) Decimal() decimal.Decimal {
	return decimal.Decimal(m)
}

func (m MoneyString) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Decimal().Round(2).StringFixed(2))
}
