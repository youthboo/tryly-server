package helper

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domainutil"
)

var requestValidator = newRequestValidator()

func newRequestValidator() *validator.Validate {
	v := validator.New()
	_ = v.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
		field := fl.Field()
		return field.Kind() == reflect.String && strings.TrimSpace(field.String()) != ""
	})
	_ = v.RegisterValidation("statuscode", func(fl validator.FieldLevel) bool {
		field := fl.Field()
		if field.Kind() != reflect.String {
			return false
		}
		value := domainutil.NormalizeStatus(field.String())
		for _, allowed := range strings.Fields(fl.Param()) {
			if value == domainutil.NormalizeStatus(allowed) {
				return true
			}
		}
		return false
	})
	return v
}

func ParseAndValidateBody(c *fiber.Ctx, out interface{}, messages map[string]string) error {
	if err := ParseBody(c, out, "invalid request payload"); err != nil {
		return err
	}
	return ValidateStruct(c, out, messages)
}

func ParseAndValidateBodyWithMessage(c *fiber.Ctx, out interface{}, messages map[string]string, parseMessage string) error {
	if err := ParseBody(c, out, parseMessage); err != nil {
		return err
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
