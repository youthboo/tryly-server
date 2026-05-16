package conversation

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

func getOptionalRoleFromContext(c *fiber.Ctx) string {
	if localValue := c.Locals("role"); localValue != nil {
		if value, ok := localValue.(string); ok {
			return strings.TrimSpace(strings.ToUpper(value))
		}
	}
	return ""
}

func unauthorized(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
}

func badRequest(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": message})
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
