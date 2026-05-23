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

func DereferenceString(ptr *string, defaultVal string) string {
	if ptr == nil {
		return defaultVal
	}
	return strings.TrimSpace(*ptr)
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
