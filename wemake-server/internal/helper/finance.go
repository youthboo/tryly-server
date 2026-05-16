package helper

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

var ErrNilDecimal = errors.New("decimal value is nil")
var ErrNegativeMoney = errors.New("money amount must be non-negative")

func MoneyDecimal(v float64) decimal.Decimal {
	d, _ := decimal.NewFromString(fmt.Sprintf("%.2f", v))
	return d
}

func MoneyDecimalFromFloat64(v float64) decimal.Decimal {
	return MoneyDecimal(v)
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

func DivideMoney(dividend decimal.Decimal, divisor decimal.Decimal) (decimal.Decimal, error) {
	if divisor.IsZero() {
		return decimal.Zero, ErrDivisionByZero
	}
	return RoundDecimal(dividend.Div(divisor)), nil
}

func ZeroMoney() decimal.Decimal {
	return decimal.Zero
}

func ParseMoney(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(s)
}

func MoneyFromFloat64(v float64) decimal.Decimal {
	return MoneyDecimal(v)
}

func MoneyFromInt(v int64) decimal.Decimal {
	return decimal.NewFromInt(v)
}

func IsMoneyZero(d decimal.Decimal) bool {
	return d.IsZero()
}

func IsMoneyEqual(a decimal.Decimal, b decimal.Decimal) bool {
	return a.Equal(b)
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

func IsMoneyGreaterOrEqual(a decimal.Decimal, b decimal.Decimal) bool {
	return a.GreaterThanOrEqual(b)
}

func PercentageOf(amount decimal.Decimal, percent decimal.Decimal) (decimal.Decimal, error) {
	if amount.IsNegative() || percent.IsNegative() {
		return decimal.Zero, ErrNegativeMoney
	}
	return DivideMoney(amount.Mul(percent), decimal.NewFromInt(100))
}

func AbsMoney(d decimal.Decimal) decimal.Decimal {
	return d.Abs()
}

func MaxMoney(a decimal.Decimal, b decimal.Decimal) decimal.Decimal {
	if a.GreaterThan(b) {
		return a
	}
	return b
}

func MinMoney(a decimal.Decimal, b decimal.Decimal) decimal.Decimal {
	if a.LessThan(b) {
		return a
	}
	return b
}

func MoneyDecimalPtr(v *float64) *decimal.Decimal {
	return MoneyDecimalPtrFromFloat64(v)
}

func MoneyDecimalPtrFromFloat64(v *float64) *decimal.Decimal {
	if v == nil {
		return nil
	}
	d := MoneyDecimal(*v)
	return &d
}

func DecimalPtrToFloatPtr(d *decimal.Decimal) *float64 {
	if d == nil {
		return nil
	}
	f := DecimalToFloat(*d)
	return &f
}

func NullableDecimalToFloat64(d *decimal.Decimal) interface{} {
	return DecimalPtrToFloatPtr(d)
}

func DerefDecimal(d *decimal.Decimal) (decimal.Decimal, error) {
	if d == nil {
		return decimal.Zero, ErrNilDecimal
	}
	return d.Round(2), nil
}

func DerefDecimalToFloat(d *decimal.Decimal) float64 {
	if d == nil {
		return 0
	}
	return DecimalToFloat(*d)
}

func MaxDecimal(a, b decimal.Decimal) decimal.Decimal {
	if a.GreaterThan(b) {
		return a
	}
	return b
}

func MinDecimal(a, b decimal.Decimal) decimal.Decimal {
	if a.LessThan(b) {
		return a
	}
	return b
}
