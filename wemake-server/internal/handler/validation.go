package handler

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
		if field.Kind() != reflect.String {
			return false
		}
		return strings.TrimSpace(field.String()) != ""
	})
	return v
}

func parseAndValidateBody(c *fiber.Ctx, out interface{}, messages map[string]string) error {
	if err := c.BodyParser(out); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid request payload")
	}
	return validateBody(c, out, messages)
}

func validateBody(c *fiber.Ctx, out interface{}, messages map[string]string) error {
	if err := requestValidator.Struct(out); err != nil {
		var message string
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			message = messages[errs[0].Field()]
		}
		if message == "" {
			message = "invalid request payload"
		}
		return jsonError(c, fiber.StatusBadRequest, message)
	}
	return nil
}

func requireBody(c *fiber.Ctx, out interface{}) error {
	if err := c.BodyParser(out); err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid request payload")
	}
	return nil
}

func validateStruct(c *fiber.Ctx, out interface{}, messages map[string]string) error {
	return validateBody(c, out, messages)
}
