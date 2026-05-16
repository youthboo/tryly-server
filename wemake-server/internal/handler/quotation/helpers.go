package quotation

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/repository"
)

var requestValidator = newRequestValidator()

type serviceErrorCase struct {
	Err     error
	Status  int
	Message string
}

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

func jsonError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
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
	if repository.IsNotFoundError(err) {
		return jsonError(c, fiber.StatusNotFound, "quotation not found")
	}
	return jsonError(c, fiber.StatusInternalServerError, fallback)
}

func badRequestCase(err error) serviceErrorCase {
	return serviceErrorCase{Err: err, Status: fiber.StatusBadRequest, Message: err.Error()}
}

func conflictCase(err error) serviceErrorCase {
	return serviceErrorCase{Err: err, Status: fiber.StatusConflict, Message: err.Error()}
}

func notFoundCase(err error, message string) serviceErrorCase {
	return serviceErrorCase{Err: err, Status: fiber.StatusNotFound, Message: message}
}
