package handler

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

	userIDStr := c.Get("X-User-ID")
	return strconv.ParseInt(userIDStr, 10, 64)
}

func getOptionalUserIDFromHeader(c *fiber.Ctx) int64 {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return 0
	}
	return userID
}

func getOptionalRoleFromContext(c *fiber.Ctx) string {
	if localValue := c.Locals("role"); localValue != nil {
		if value, ok := localValue.(string); ok {
			return strings.TrimSpace(strings.ToUpper(value))
		}
	}
	return ""
}

func jsonError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxIntQuery(v, min int) int {
	if v < min {
		return min
	}
	return v
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

func pageLimit(c *fiber.Ctx, defaultLimit int) (int, int) {
	return maxIntQuery(c.QueryInt("page", 1), 1), clampInt(c.QueryInt("limit", defaultLimit), 1, 100)
}

func limitOffset(c *fiber.Ctx, defaultLimit int) (int, int) {
	return clampInt(c.QueryInt("limit", defaultLimit), 1, 100), maxIntQuery(c.QueryInt("offset", 0), 0)
}
