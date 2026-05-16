package service

import (
	"fmt"
	"time"

	"github.com/yourusername/wemake/internal/domainutil"
)

func derefInt64(v *int64) int64 {
	return domainutil.Int64Value(v)
}

func derefInt(v *int) int {
	return domainutil.IntValue(v)
}

func derefFloat64(v *float64) float64 {
	return domainutil.Float64Value(v)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func formatThaiShortDate(t time.Time) string {
	months := []string{"ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."}
	return fmt.Sprintf("%d %s %02d", t.Day(), months[int(t.Month())-1], (t.Year()+543)%100)
}

func nullableBOQInt(v *int) interface{} {
	return domainutil.NullableInt(v)
}

func nullableBOQInt64(v *int64) interface{} {
	return domainutil.NullableInt64(v)
}

func nullableBOQString(v *string) interface{} {
	return domainutil.NullableString(v)
}
