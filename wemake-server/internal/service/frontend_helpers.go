package service

import (
	"github.com/yourusername/wemake/internal/helper"
	"github.com/yourusername/wemake/internal/repository"
)

func parseFactoryID(id string) int64 {
	return helper.ParseFactoryID(id)
}

func parseRFQID(id string) int64 {
	return helper.ParseRFQID(id)
}

func slugifyCategory(name string) string {
	return helper.SlugifyCategory(name)
}

func factoryImageURL(factoryID int64) string {
	return helper.FactoryImageURL(factoryID)
}

func avatarURL(name string) string {
	return helper.AvatarURL(name)
}

func formatLeadTimeRange(avg float64) string {
	return helper.FormatLeadTimeRange(avg)
}

func dateDaysAgo(days int) string {
	return helper.DateDaysAgo(days)
}

func dateDaysFromNow(days int) string {
	return helper.DateDaysFromNow(days)
}

func relativeThaiTime(date string) string {
	return helper.RelativeThaiTime(date)
}

func relativeThaiTimeFromISO(date string) string {
	return helper.RelativeThaiTimeFromISO(date)
}

func optionalString(value string) *string {
	return helper.OptionalString(value)
}

func lastMessageTime(items []repository.FrontendMessageRow) string {
	if len(items) == 0 {
		return ""
	}
	return items[len(items)-1].CreatedAt
}
