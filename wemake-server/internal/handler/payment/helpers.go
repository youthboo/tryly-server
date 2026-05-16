package payment

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

func badRequest(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": message})
}

func jsonError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}

func requireBody(c *fiber.Ctx, out interface{}) error {
	if err := c.BodyParser(out); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request payload"})
	}
	return nil
}

func parsePositiveInt64Param(c *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(c.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return value, nil
}

func parseRequiredDateValue(value string, field string) (time.Time, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return time.Time{}, fiber.NewError(fiber.StatusBadRequest, field+" is required")
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
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
	if errors.Is(err, sql.ErrNoRows) || repository.IsNotFoundError(err) {
		return jsonError(c, fiber.StatusNotFound, "not found")
	}
	return jsonError(c, fiber.StatusInternalServerError, fallback)
}

func notFoundCase(err error, message string) serviceErrorCase {
	return serviceErrorCase{Err: err, Status: fiber.StatusNotFound, Message: message}
}
