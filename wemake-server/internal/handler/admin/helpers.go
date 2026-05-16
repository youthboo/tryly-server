package admin

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/repository"
)

var errNotFound = sql.ErrNoRows

type serviceErrorCase struct {
	Err     error
	Status  int
	Message string
}

func getUserIDFromHeader(c *fiber.Ctx) (int64, error) {
	if localValue := c.Locals("user_id"); localValue != nil {
		switch value := localValue.(type) {
		case int64:
			return value, nil
		case int:
			return int64(value), nil
		case string:
			return strconv.ParseInt(value, 10, 64)
		}
	}
	return strconv.ParseInt(c.Get("X-User-ID"), 10, 64)
}

func jsonError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}

func unauthorized(c *fiber.Ctx) error {
	return jsonError(c, fiber.StatusUnauthorized, "unauthorized")
}

func badRequest(c *fiber.Ctx, message string) error {
	return jsonError(c, fiber.StatusBadRequest, message)
}

func requireBody(c *fiber.Ctx, out interface{}) error {
	if err := c.BodyParser(out); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid request payload")
	}
	return nil
}

func writeServiceError(c *fiber.Ctx, err error, fallback string, cases ...serviceErrorCase) error {
	for _, item := range cases {
		if item.Err != nil && errors.Is(err, item.Err) {
			return jsonError(c, item.Status, item.Message)
		}
	}
	if errors.Is(err, domain.ErrForbidden) {
		return jsonError(c, fiber.StatusForbidden, "forbidden")
	}
	if repository.IsNotFoundError(err) {
		return jsonError(c, fiber.StatusNotFound, "not found")
	}
	return jsonError(c, fiber.StatusInternalServerError, fallback)
}

func notFoundCase(err error, message string) serviceErrorCase {
	return serviceErrorCase{Err: err, Status: fiber.StatusNotFound, Message: message}
}

func parsePositiveInt64Param(c *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(c.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return value, nil
}

func parseOptionalPositiveInt64Query(c *fiber.Ctx, name string) (*int64, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return &value, nil
}

func parseOptionalDateQuery(c *fiber.Ctx, name string) (*time.Time, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}
	value, err := parseDate(raw, name)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func parseDate(raw string, name string) (time.Time, error) {
	value, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, fiber.NewError(fiber.StatusBadRequest, name+" must be YYYY-MM-DD")
	}
	return value, nil
}

func parseOptionalRFC3339Value(raw *string, name string) (*time.Time, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	value, err := time.Parse(time.RFC3339, strings.TrimSpace(*raw))
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, name+" must be RFC3339")
	}
	return &value, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func normalizePageSize(size int) int {
	if size <= 0 {
		return 20
	}
	if size > 100 {
		return 100
	}
	return size
}

func limitOffset(c *fiber.Ctx, defaultLimit int) (int, int) {
	return clampInt(c.QueryInt("limit", defaultLimit), 1, 100), maxInt(c.QueryInt("offset", 0), 0)
}
