package helper

import (
	"fmt"
	"strings"
	"time"

	"github.com/yourusername/wemake/internal/domainutil"
)

var ThailandLocation = time.FixedZone("Asia/Bangkok", 7*60*60)

func RoundCurrency(v float64) float64 {
	return domainutil.RoundMoney(v)
}

func RoundMoney(v float64) float64 {
	return domainutil.RoundMoney(v)
}

func PercentOf(amount, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return RoundCurrency((amount / total) * 100)
}

func DerefString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func DerefInt64(v *int64) int64 {
	return domainutil.Int64Value(v)
}

func DerefInt(v *int) int {
	return domainutil.IntValue(v)
}

func DerefFloat64(v *float64) float64 {
	return domainutil.Float64Value(v)
}

func NullableInt(v *int) interface{} {
	return domainutil.NullableInt(v)
}

func NullableInt64(v *int64) interface{} {
	return domainutil.NullableInt64(v)
}

func NullableString(v *string) interface{} {
	return domainutil.NullableString(v)
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func FormatThaiShortDate(t time.Time) string {
	months := []string{"ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."}
	return fmt.Sprintf("%d %s %02d", t.Day(), months[int(t.Month())-1], (t.Year()+543)%100)
}

func NormalizeString(s string) string {
	return strings.TrimSpace(s)
}

func NormalizeEmail(s string) string {
	return domainutil.NormalizeLower(s)
}

func NormalizeRole(s string) string {
	return domainutil.NormalizeStatus(s)
}

func NormalizePhone(s string) string {
	return strings.TrimSpace(s)
}

func NormalizeName(s string) string {
	return strings.TrimSpace(s)
}

func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

func DereferenceString(ptr *string, defaultVal string) string {
	if ptr == nil {
		return defaultVal
	}
	return strings.TrimSpace(*ptr)
}

func DereferenceInt64(ptr *int64, defaultVal int64) int64 {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func DereferenceInt(ptr *int, defaultVal int) int {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func AssignIfNotNil[T any](target *T, ptr *T) {
	if ptr != nil {
		*target = *ptr
	}
}

func AssignStringIfNotNil(target *string, ptr *string) {
	if ptr != nil {
		trimmed := strings.TrimSpace(*ptr)
		*target = trimmed
	}
}

func MergePointerString(target **string, ptr *string) {
	if ptr != nil {
		trimmed := strings.TrimSpace(*ptr)
		*target = &trimmed
	}
}

func MergePointerValue[T any](target **T, ptr *T) {
	if ptr != nil {
		*target = ptr
	}
}

func ValidatePointerInt64(ptr *int64, fieldName string) error {
	if ptr == nil || *ptr <= 0 {
		return fmt.Errorf("%s is required and must be positive", fieldName)
	}
	return nil
}

func ValidatePointerInt(ptr *int, fieldName string) error {
	if ptr == nil || *ptr <= 0 {
		return fmt.Errorf("%s is required and must be positive", fieldName)
	}
	return nil
}

func ValidatePointerString(ptr *string, fieldName string) error {
	if ptr == nil || strings.TrimSpace(*ptr) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

func ValidatePointerBool(ptr *bool, fieldName string) error {
	if ptr == nil {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}
