package helper

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func ParseFactoryID(id string) int64 {
	value := strings.TrimPrefix(id, "f")
	result, _ := strconv.ParseInt(value, 10, 64)
	return result
}

func ParseRFQID(id string) int64 {
	value := strings.TrimPrefix(id, "rfq")
	result, _ := strconv.ParseInt(value, 10, 64)
	return result
}

func SlugifyCategory(name string) string {
	switch name {
	case "อาหารสัตว์":
		return "pet_food"
	case "อาหารเสริม":
		return "supplements"
	case "ของเล่นสัตว์เลี้ยง":
		return "pet_toys"
	case "สายจูง อุปกรณ์", "อุปกรณ์สัตว์เลี้ยง":
		return "leash_equipment"
	case "เสื้อผ้าสัตว์เลี้ยง":
		return "pet_clothes"
	default:
		return "other"
	}
}

func FactoryImageURL(factoryID int64) string {
	images := []string{
		"https://images.unsplash.com/photo-1579784340946-55a7bbd51d57?w=400&h=250&fit=crop",
		"https://images.unsplash.com/photo-1684259499086-93cb3e555803?w=400&h=250&fit=crop",
		"https://images.unsplash.com/photo-1587300003388-59208cc962cb?w=400&h=250&fit=crop",
		"https://images.unsplash.com/photo-1607082348824-0a96f2a4b9da?w=400&h=250&fit=crop",
		"https://images.unsplash.com/photo-1471864190281-a93a3070b6de?w=400&h=250&fit=crop",
		"https://images.unsplash.com/photo-1517849845537-4d257902454a?w=400&h=250&fit=crop",
	}
	return images[(factoryID-1)%int64(len(images))]
}

func AvatarURL(name string) string {
	return "https://ui-avatars.com/api/?background=EDE9FF&color=6C47FF&name=" + url.QueryEscape(name)
}

func FormatLeadTimeRange(avg float64) string {
	if avg <= 0 {
		return ""
	}
	base := int(avg + 0.5)
	return fmt.Sprintf("%d-%d วัน", maxInt(base-2, 1), base+2)
}

func DateDaysAgo(days int) string {
	return time.Now().AddDate(0, 0, -days).Format("2006-01-02")
}

func DateDaysFromNow(days int) string {
	return time.Now().AddDate(0, 0, days).Format("2006-01-02")
}

func RelativeThaiTime(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	diff := time.Since(t)
	if diff.Hours() < 24 {
		return "วันนี้"
	}
	if diff.Hours() < 48 {
		return "เมื่อวาน"
	}
	return fmt.Sprintf("%d วันที่แล้ว", int(diff.Hours()/24))
}

func RelativeThaiTimeFromISO(date string) string {
	t, err := time.Parse("2006-01-02T15:04:05", date)
	if err != nil {
		return date
	}
	diff := time.Since(t)
	if diff.Minutes() < 60 {
		return fmt.Sprintf("%d นาทีที่แล้ว", int(diff.Minutes()))
	}
	if diff.Hours() < 24 {
		return fmt.Sprintf("%d ชั่วโมงที่แล้ว", int(diff.Hours()))
	}
	if diff.Hours() < 48 {
		return "เมื่อวาน"
	}
	return fmt.Sprintf("%d วันที่แล้ว", int(diff.Hours()/24))
}

func OptionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	v := value
	return &v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
