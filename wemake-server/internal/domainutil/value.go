package domainutil

import (
	"math"
	"strings"
	"time"
)

func RoundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func StringValue(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func IntValue(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func Int64Value(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func Float64Value(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func BoolValue(v *bool) bool {
	return v != nil && *v
}

func NullableString(v *string) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func NullableInt(v *int) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func NullableInt64(v *int64) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func NullablePositiveInt64(v int64) interface{} {
	if v <= 0 {
		return nil
	}
	return v
}

func NullableFloat64(v *float64) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func NullableTime(v *time.Time) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func NormalizeStatus(v string) string {
	return strings.ToUpper(strings.TrimSpace(v))
}

func NormalizeUpperOrDefault(v string, fallback string) string {
	normalized := strings.ToUpper(strings.TrimSpace(v))
	if normalized == "" {
		return fallback
	}
	return normalized
}
