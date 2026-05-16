package helper

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
)

var ErrNotFound = sql.ErrNoRows

type ServiceErrorCase struct {
	Err     error
	Status  int
	Message string
}

type ErrorResponse struct {
	Status int
	Body   fiber.Map
}

type ErrorResponseBuilder func(error) ErrorResponse

func UserIDFromHeader(c *fiber.Ctx) (int64, error) {
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

func OptionalUserIDFromHeader(c *fiber.Ctx) int64 {
	userID, err := UserIDFromHeader(c)
	if err != nil {
		return 0
	}
	return userID
}

func OptionalRoleFromContext(c *fiber.Ctx) string {
	if localValue := c.Locals("role"); localValue != nil {
		if value, ok := localValue.(string); ok {
			return strings.TrimSpace(strings.ToUpper(value))
		}
	}
	return ""
}

func JSONError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}

func Unauthorized(c *fiber.Ctx) error {
	return JSONError(c, fiber.StatusUnauthorized, "unauthorized")
}

func BadRequest(c *fiber.Ctx, message string) error {
	return JSONError(c, fiber.StatusBadRequest, message)
}

func RequireBody(c *fiber.Ctx, out interface{}) error {
	return ParseBody(c, out, "invalid request payload")
}

func ParseBody(c *fiber.Ctx, out interface{}, message string) error {
	if err := c.BodyParser(out); err != nil {
		if message == "" {
			message = "invalid request payload"
		}
		return JSONError(c, fiber.StatusBadRequest, message)
	}
	return nil
}

func ParseJSONBody(c *fiber.Ctx, out interface{}, message string) error {
	if err := UnmarshalJSON(c.Body(), out); err != nil {
		if message == "" {
			message = "invalid request payload"
		}
		return JSONError(c, fiber.StatusBadRequest, message)
	}
	return nil
}

func UnmarshalJSON(data []byte, out interface{}) error {
	return json.Unmarshal(data, out)
}

func ParsePositiveInt64Param(c *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(c.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return value, nil
}

func ParsePositiveInt64Value(raw string, name string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return value, nil
}

func ParseOptionalPositiveInt64Query(c *fiber.Ctx, name string) (*int64, error) {
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

func ParseOptionalPositiveInt64Value(raw string, name string) (*int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := ParsePositiveInt64Value(raw, name)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func ParseRequiredPositiveInt64Query(c *fiber.Ctx, name string) (int64, error) {
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

func ParseOptionalDateQuery(c *fiber.Ctx, name string) (*time.Time, error) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil, nil
	}
	value, err := ParseDate(raw, name)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func ParseRequiredDateValue(raw string, name string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fiber.NewError(fiber.StatusBadRequest, name+" is required")
	}
	return ParseDate(raw, name)
}

func ParseDate(raw string, name string) (time.Time, error) {
	value, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, fiber.NewError(fiber.StatusBadRequest, name+" must be YYYY-MM-DD")
	}
	return value, nil
}

func ParseOptionalRFC3339Value(raw *string, name string) (*time.Time, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	value, err := time.Parse(time.RFC3339, strings.TrimSpace(*raw))
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, name+" must be RFC3339")
	}
	return &value, nil
}

func WriteServiceError(c *fiber.Ctx, err error, fallback string, cases ...ServiceErrorCase) error {
	return WriteServiceErrorWithNotFound(c, err, fallback, "not found", cases...)
}

func WriteServiceErrorWithNotFound(c *fiber.Ctx, err error, fallback string, notFoundMessage string, cases ...ServiceErrorCase) error {
	for _, item := range cases {
		if item.Err != nil && errors.Is(err, item.Err) {
			return JSONError(c, item.Status, item.Message)
		}
	}
	if errors.Is(err, domain.ErrForbidden) {
		return JSONError(c, fiber.StatusForbidden, "forbidden")
	}
	if errors.Is(err, sql.ErrNoRows) {
		if notFoundMessage == "" {
			notFoundMessage = "not found"
		}
		return JSONError(c, fiber.StatusNotFound, notFoundMessage)
	}
	return JSONError(c, fiber.StatusInternalServerError, fallback)
}

func MapServiceError(c *fiber.Ctx, err error, fallback ErrorResponse, responses map[error]ErrorResponse) error {
	for target, response := range responses {
		if target != nil && errors.Is(err, target) {
			return WriteErrorResponse(c, response)
		}
	}
	return WriteErrorResponse(c, fallback)
}

func MapServiceErrorFunc(c *fiber.Ctx, err error, fallback ErrorResponse, responses map[error]ErrorResponseBuilder) error {
	for target, buildResponse := range responses {
		if target != nil && errors.Is(err, target) {
			if buildResponse == nil {
				return WriteErrorResponse(c, fallback)
			}
			return WriteErrorResponse(c, buildResponse(err))
		}
	}
	return WriteErrorResponse(c, fallback)
}

func WriteErrorResponse(c *fiber.Ctx, response ErrorResponse) error {
	status := response.Status
	if status == 0 {
		status = fiber.StatusInternalServerError
	}
	body := response.Body
	if body == nil {
		body = fiber.Map{"error": "internal server error"}
	}
	return c.Status(status).JSON(body)
}

func ErrorMessage(status int, message string) ErrorResponse {
	return ErrorResponse{Status: status, Body: fiber.Map{"error": message}}
}

func ErrorBody(status int, body fiber.Map) ErrorResponse {
	return ErrorResponse{Status: status, Body: body}
}

func BadRequestCase(err error) ServiceErrorCase {
	return ServiceErrorCase{Err: err, Status: fiber.StatusBadRequest, Message: err.Error()}
}

func ConflictCase(err error) ServiceErrorCase {
	return ServiceErrorCase{Err: err, Status: fiber.StatusConflict, Message: err.Error()}
}

func NotFoundCase(err error, message string) ServiceErrorCase {
	return ServiceErrorCase{Err: err, Status: fiber.StatusNotFound, Message: message}
}

func UnprocessableCase(err error) ServiceErrorCase {
	return ServiceErrorCase{Err: err, Status: fiber.StatusUnprocessableEntity, Message: err.Error()}
}

func ForbiddenCase(err error) ServiceErrorCase {
	return ServiceErrorCase{Err: err, Status: fiber.StatusForbidden, Message: err.Error()}
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MaxIntQuery(v, min int) int {
	if v < min {
		return min
	}
	return v
}

func ClampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func NormalizePageSize(size int) int {
	if size <= 0 {
		return 20
	}
	if size > 100 {
		return 100
	}
	return size
}

func PageLimit(c *fiber.Ctx, defaultLimit int) (int, int) {
	return MaxIntQuery(c.QueryInt("page", 1), 1), ClampInt(c.QueryInt("limit", defaultLimit), 1, 100)
}

func LimitOffset(c *fiber.Ctx, defaultLimit int) (int, int) {
	return ClampInt(c.QueryInt("limit", defaultLimit), 1, 100), MaxInt(c.QueryInt("offset", 0), 0)
}
