package service

import (
	"fmt"
	"time"
)

func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func derefInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func derefFloat64(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
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
	if v == nil {
		return nil
	}
	return *v
}

func nullableBOQInt64(v *int64) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func nullableBOQString(v *string) interface{} {
	if v == nil {
		return nil
	}
	return *v
}
