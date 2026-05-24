package helper

import (
	"github.com/shopspring/decimal"
)

func MoneyDecimal(v float64) decimal.Decimal {
	return decimal.NewFromFloat(v).Round(2)
}

func DecimalToFloat(d decimal.Decimal) float64 {
	f, _ := d.Float64()
	return f
}

func RoundDecimal(d decimal.Decimal) decimal.Decimal {
	return d.Round(2)
}

func AddMoney(amounts ...decimal.Decimal) decimal.Decimal {
	result := decimal.Zero
	for _, amount := range amounts {
		result = result.Add(amount)
	}
	return RoundDecimal(result)
}

func SubtractMoney(minuend decimal.Decimal, subtrahend decimal.Decimal) decimal.Decimal {
	return RoundDecimal(minuend.Sub(subtrahend))
}

func MultiplyMoney(a decimal.Decimal, b decimal.Decimal) decimal.Decimal {
	return RoundDecimal(a.Mul(b))
}

func ZeroMoney() decimal.Decimal {
	return decimal.Zero
}

func MoneyFromInt(v int64) decimal.Decimal {
	return decimal.NewFromInt(v)
}

func IsMoneyZero(d decimal.Decimal) bool {
	return d.IsZero()
}

func IsMoneyLess(a decimal.Decimal, b decimal.Decimal) bool {
	return a.LessThan(b)
}

func IsMoneyLessOrEqual(a decimal.Decimal, b decimal.Decimal) bool {
	return a.LessThanOrEqual(b)
}

func IsMoneyGreater(a decimal.Decimal, b decimal.Decimal) bool {
	return a.GreaterThan(b)
}

func MoneyDecimalPtrFromFloat64(v *float64) *decimal.Decimal {
	if v == nil {
		return nil
	}
	d := MoneyDecimal(*v)
	return &d
}

func DerefDecimalToFloat(d *decimal.Decimal) float64 {
	if d == nil {
		return 0
	}
	return DecimalToFloat(*d)
}
