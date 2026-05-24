package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
)

// RequireAuth middleware ตรวจสอบว่า user_id มีค่าและถูกต้อง
// ใช้หลังจาก AuthContext middleware
func RequireAuth(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.JSONError(c, fiber.StatusBadRequest, "invalid user ID")
	}
	if userID <= 0 {
		return helper.JSONError(c, fiber.StatusBadRequest, "invalid user ID")
	}
	// user_id ถูกต้องแล้ว ส่งต่อไป
	return c.Next()
}
