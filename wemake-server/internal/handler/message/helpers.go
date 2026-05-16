package message

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

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

func parseRequiredPositiveInt64Query(c *fiber.Ctx, name string) (int64, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return 0, fiber.NewError(fiber.StatusBadRequest, name+" is required")
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, name+" must be a positive integer")
	}
	return value, nil
}
