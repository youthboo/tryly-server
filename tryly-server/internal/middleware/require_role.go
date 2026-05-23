package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	authservice "github.com/yourusername/wemake/internal/service/auth"
	"github.com/yourusername/wemake/internal/domainutil"
)

// RequireRole middleware ตรวจสอบว่า user มี role ที่อนุญาต
// Usage: app.Use(RequireRole(authService, "admin", "super_admin"))
func RequireRole(auth *authservice.AuthService, roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[domainutil.NormalizeStatus(role)] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		userID, err := helper.UserIDFromHeader(c)
		if err != nil || userID <= 0 {
			return helper.WriteAPIError(c, helper.UnauthorizedAPIError("UNAUTHORIZED", "invalid user"))
		}

		user, err := auth.GetUserByID(userID)
		if err != nil {
			return helper.WriteAPIError(c, helper.UnauthorizedAPIError("USER_NOT_FOUND", "user not found"))
		}

		if _, ok := allowed[domainutil.NormalizeStatus(user.Role)]; !ok {
			return helper.WriteAPIError(c, helper.ForbiddenAPIError("INSUFFICIENT_PERMISSION", "insufficient permissions"))
		}

		// Store user in context สำหรับ handlers ใช้ต่อ
		c.Locals("user", user)
		return c.Next()
	}
}

// RequireUserAndRole middleware ตรวจสอบ user + role + set user in context
// More convenient when you need both user object and role validation
func RequireUserAndRole(auth *authservice.AuthService, requiredRoles ...string) fiber.Handler {
	allowedRoles := make(map[string]struct{}, len(requiredRoles))
	for _, role := range requiredRoles {
		allowedRoles[domainutil.NormalizeStatus(role)] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		// Extract user ID
		userID, err := helper.UserIDFromHeader(c)
		if err != nil || userID <= 0 {
			return helper.WriteAPIError(c, helper.UnauthorizedAPIError("INVALID_USER", "invalid user ID"))
		}

		// Fetch user
		user, err := auth.GetUserByID(userID)
		if err != nil {
			return helper.WriteAPIError(c, helper.UnauthorizedAPIError("USER_NOT_FOUND", "user not found"))
		}

		// Check role
		userRole := domainutil.NormalizeStatus(user.Role)
		if _, ok := allowedRoles[userRole]; !ok {
			return helper.WriteAPIError(c, helper.ForbiddenAPIError("INSUFFICIENT_ROLE", "user role not allowed"))
		}

		// Store user in context for downstream handlers
		c.Locals("user", user)
		c.Locals("user_role", userRole)

		return c.Next()
	}
}

// OptionalRequireRole middleware ตรวจสอบ role ถ้า user authenticated
// If not authenticated, just continue (useful for mixed public/private endpoints)
func OptionalRequireRole(auth *authservice.AuthService, requiredRoles ...string) fiber.Handler {
	allowedRoles := make(map[string]struct{}, len(requiredRoles))
	for _, role := range requiredRoles {
		allowedRoles[domainutil.NormalizeStatus(role)] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		userID, err := helper.UserIDFromHeader(c)
		if err != nil || userID <= 0 {
			// User not authenticated, continue
			return c.Next()
		}

		user, err := auth.GetUserByID(userID)
		if err != nil {
			// User not found, continue without auth
			return c.Next()
		}

		// Check if role matches
		if _, ok := allowedRoles[domainutil.NormalizeStatus(user.Role)]; ok {
			c.Locals("user", user)
		}

		return c.Next()
	}
}
