package factory

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var requestValidator = newRequestValidator()

func newRequestValidator() *validator.Validate {
	v := validator.New()
	_ = v.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
		field := fl.Field()
		return field.Kind() == reflect.String && strings.TrimSpace(field.String()) != ""
	})
	return v
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

func parseAndValidateBody(c *fiber.Ctx, out interface{}, messages map[string]string) error {
	if err := c.BodyParser(out); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request payload"})
	}
	return validateStruct(c, out, messages)
}

func validateStruct(c *fiber.Ctx, out interface{}, messages map[string]string) error {
	if err := requestValidator.Struct(out); err != nil {
		message := "invalid request payload"
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			if custom := messages[errs[0].Field()]; custom != "" {
				message = custom
			}
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": message})
	}
	return nil
}
