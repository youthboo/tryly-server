package admin

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	adminservice "github.com/yourusername/wemake/internal/service/admin"
)

type AdminDashboardHandler struct {
	service *adminservice.AdminDashboardService
}

func NewAdminDashboardHandler(service *adminservice.AdminDashboardService) *AdminDashboardHandler {
	return &AdminDashboardHandler{service: service}
}

func (h *AdminDashboardHandler) GetSummary(c *fiber.Ctx) error {
	period := helper.QueryString(c, "period")
	if period == "" {
		period = "month"
	}
	from, to, err := parseAdminPeriod(period, helper.QueryString(c, "date_from"), helper.QueryString(c, "date_to"))
	if err != nil {
		return helper.BadRequestError(c, err.Error())
	}
	item, err := h.service.GetSummary(from, to, period)
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch dashboard summary")
	}
	return c.JSON(item)
}

func (h *AdminDashboardHandler) GetRevenueChart(c *fiber.Ctx) error {
	granularity := normalizeDashboardGranularity(helper.QueryString(c, "granularity"))
	from, to, err := parseAdminRange(helper.QueryString(c, "date_from"), helper.QueryString(c, "date_to"))
	if err != nil {
		return helper.BadRequestError(c, err.Error())
	}
	items, err := h.service.GetRevenueChart(from, to, granularity)
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch revenue chart")
	}
	return c.JSON(fiber.Map{"granularity": granularity, "data": items})
}

func (h *AdminDashboardHandler) GetTopFactories(c *fiber.Ctx) error {
	query := helper.QueryParams(c)
	from, to, err := parseAdminRange(helper.QueryString(c, "date_from"), helper.QueryString(c, "date_to"))
	if err != nil {
		return helper.BadRequestError(c, err.Error())
	}
	items, err := h.service.GetTopFactories(from, to, query.Int("limit", 10))
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch top factories")
	}
	return c.JSON(fiber.Map{"data": items})
}

func parseAdminPeriod(period, rawFrom, rawTo string) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	switch period {
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return start, start.Add(24 * time.Hour), nil
	case "week":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return start.AddDate(0, 0, -6), start.Add(24 * time.Hour), nil
	case "month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return start, start.AddDate(0, 1, 0), nil
	case "quarter":
		month := ((int(now.Month())-1)/3)*3 + 1
		start := time.Date(now.Year(), time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		return start, start.AddDate(0, 3, 0), nil
	case "year":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		return start, start.AddDate(1, 0, 0), nil
	case "custom":
		if rawFrom == "" || rawTo == "" {
			return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "date_from and date_to are required")
		}
		from, err := helper.ParseDate(rawFrom, "date_from")
		if err != nil {
			return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "date_from must be YYYY-MM-DD")
		}
		to, err := helper.ParseDate(rawTo, "date_to")
		if err != nil {
			return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "date_to must be YYYY-MM-DD")
		}
		return from, to.Add(24 * time.Hour), nil
	default:
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "invalid period")
	}
}

func parseAdminRange(rawFrom, rawTo string) (time.Time, time.Time, error) {
	if rawFrom == "" && rawTo == "" {
		return parseAdminPeriod("month", "", "")
	}
	if rawFrom == "" || rawTo == "" {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "date_from and date_to must be provided together")
	}
	return parseAdminPeriod("custom", rawFrom, rawTo)
}

func normalizeDashboardGranularity(granularity string) string {
	switch granularity {
	case "week", "month":
		return granularity
	default:
		return "day"
	}
}
