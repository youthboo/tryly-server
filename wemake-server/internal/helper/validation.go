package helper

import (
	"reflect"
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

func ParseAndValidateBody(c *fiber.Ctx, out interface{}, messages map[string]string) error {
	if err := c.BodyParser(out); err != nil {
		return JSONError(c, fiber.StatusBadRequest, "invalid request payload")
	}
	return ValidateStruct(c, out, messages)
}

func ValidateStruct(c *fiber.Ctx, out interface{}, messages map[string]string) error {
	if err := requestValidator.Struct(out); err != nil {
		message := "invalid request payload"
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			if custom := messages[errs[0].Field()]; custom != "" {
				message = custom
			}
		}
		return JSONError(c, fiber.StatusBadRequest, message)
	}
	return nil
}
