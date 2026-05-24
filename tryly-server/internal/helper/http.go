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
	"github.com/yourusername/wemake/internal/domainutil"
)

var (
	ErrNotFound            = sql.ErrNoRows
	ErrRoleRequired        = errors.New("role required")
	ErrFactoryRoleRequired = errors.New("factory role required")
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
	MinPageSize     = 1
)

type ServiceErrorCase struct {
	Err     error
	Status  int
	Message string
}

type ErrorResponse struct {
	Status int
	Body   fiber.Map
}

// APIError เป็น standardized error response ที่ consistent ทั่วทั้ง API
type APIError struct {
	ErrorCode string                 `json:"error_code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Status    int                    `json:"-"` // ไม่ serialize ไปใน JSON
}

type ErrorResponseBuilder func(error) ErrorResponse

type UserLoader interface {
	GetUserByID(userID int64) (*domain.User, error)
}

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

func OptionalActorID(c *fiber.Ctx) int64 {
	return OptionalUserIDFromHeader(c)
}

func OptionalRoleFromContext(c *fiber.Ctx) string {
	if localValue := c.Locals("role"); localValue != nil {
		if value, ok := localValue.(string); ok {
			return domainutil.NormalizeStatus(value)
		}
	}
	return ""
}

func RequireUserID(c *fiber.Ctx) (int64, error) {
	userID, err := UserIDFromHeader(c)
	if err != nil {
		return 0, BadRequest(c, "invalid X-User-ID header")
	}
	return userID, nil
}

func RequireAuthenticatedUserID(c *fiber.Ctx) (int64, error) {
	userID, err := UserIDFromHeader(c)
	if err != nil {
		return 0, Unauthorized(c)
	}
	return userID, nil
}

func RequireUser(c *fiber.Ctx, loader UserLoader) (int64, *domain.User, error) {
	userID, err := RequireUserID(c)
	if err != nil {
		return 0, nil, err
	}
	user, err := loader.GetUserByID(userID)
	if err != nil {
		return 0, nil, JSONError(c, fiber.StatusUnauthorized, "user not found")
	}
	return userID, user, nil
}

func RequireFactoryUser(c *fiber.Ctx, loader UserLoader) (int64, *domain.User, error) {
	userID, user, err := RequireUser(c, loader)
	if err != nil {
		return 0, nil, err
	}
	if err := RequireFactoryRole(user); err != nil {
		return 0, nil, JSONError(c, fiber.StatusForbidden, "factory role required")
	}
	return userID, user, nil
}

func RequireAPIUserID(c *fiber.Ctx, invalidUserError *APIError) (int64, error) {
	userID, err := UserIDFromHeader(c)
	if err != nil {
		if invalidUserError == nil {
			invalidUserError = BadRequestAPIError("INVALID_USER_CONTEXT", "invalid user context")
		}
		return 0, WriteAPIError(c, invalidUserError)
	}
	return userID, nil
}

func RequireAPIUser(c *fiber.Ctx, loader UserLoader, invalidUserError *APIError, notFoundError *APIError) (int64, *domain.User, error) {
	userID, err := RequireAPIUserID(c, invalidUserError)
	if err != nil {
		return 0, nil, err
	}
	user, err := loader.GetUserByID(userID)
	if err != nil {
		if notFoundError == nil {
			notFoundError = UnauthorizedAPIError("USER_NOT_FOUND", "user not found")
		}
		return 0, nil, WriteAPIError(c, notFoundError)
	}
	return userID, user, nil
}

func RequireAPIFactoryUser(c *fiber.Ctx, loader UserLoader, invalidUserError *APIError, notFoundError *APIError, forbiddenError *APIError) (int64, *domain.User, error) {
	userID, user, err := RequireAPIUser(c, loader, invalidUserError, notFoundError)
	if err != nil {
		return 0, nil, err
	}
	if err := RequireFactoryRole(user); err != nil {
		if forbiddenError == nil {
			forbiddenError = ForbiddenAPIError("FACTORY_ROLE_REQUIRED", "factory role required")
		}
		return 0, nil, WriteAPIError(c, forbiddenError)
	}
	return userID, user, nil
}

func RequireRole(user *domain.User, allowed ...string) error {
	if user == nil || len(allowed) == 0 {
		return ErrRoleRequired
	}
	role := domainutil.NormalizeStatus(user.Role)
	for _, item := range allowed {
		if role == domainutil.NormalizeStatus(item) {
			return nil
		}
	}
	return ErrRoleRequired
}

func RequireFactoryRole(user *domain.User) error {
	if err := RequireRole(user, domain.RoleFactory); err != nil {
		return ErrFactoryRoleRequired
	}
	return nil
}

func JSONError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(ErrorMap(message))
}

func ErrorMap(message string) fiber.Map {
	return fiber.Map{"error": message}
}

func Unauthorized(c *fiber.Ctx) error {
	return UnauthorizedError(c, "unauthorized")
}

func BadRequest(c *fiber.Ctx, message string) error {
	return BadRequestError(c, message)
}

func JSONInternal(c *fiber.Ctx, message string) error {
	return InternalServerError(c, message)
}

func BadRequestError(c *fiber.Ctx, message string) error {
	return JSONError(c, fiber.StatusBadRequest, message)
}

func InternalServerError(c *fiber.Ctx, message string) error {
	return JSONError(c, fiber.StatusInternalServerError, message)
}

func UnauthorizedError(c *fiber.Ctx, message string) error {
	return JSONError(c, fiber.StatusUnauthorized, message)
}

func ForbiddenError(c *fiber.Ctx, message string) error {
	return JSONError(c, fiber.StatusForbidden, message)
}

func PaginatedResponse(c *fiber.Ctx, data interface{}, page, pageSize, total int) error {
	return c.JSON(fiber.Map{
		"data": data,
		"pagination": domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
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
	return RequireInt64Param(c, name)
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

func OptionalAPIPositiveInt64Query(c *fiber.Ctx, name string, invalidQueryError *APIError) (*int64, error) {
	value, err := ParseOptionalPositiveInt64Query(c, name)
	if err != nil {
		if invalidQueryError == nil {
			invalidQueryError = BadRequestAPIError("INVALID_QUERY_PARAM", "invalid "+name)
		}
		return nil, WriteAPIError(c, invalidQueryError)
	}
	return value, nil
}

func QueryString(c *fiber.Ctx, name string) string {
	return strings.TrimSpace(c.Query(name))
}

func ParamString(c *fiber.Ctx, name string) string {
	return strings.TrimSpace(c.Params(name))
}

func HeaderString(c *fiber.Ctx, name string) string {
	return strings.TrimSpace(c.Get(name))
}

func QuerySelfOrID(c *fiber.Ctx, name string) (*int64, error) {
	raw := QueryString(c, name)
	if raw == "" {
		return nil, nil
	}
	if strings.EqualFold(raw, "me") {
		userID, err := RequireAuthenticatedUserID(c)
		if err != nil {
			return nil, err
		}
		return &userID, nil
	}
	value, err := ParsePositiveInt64Value(raw, name)
	if err != nil {
		return nil, BadRequest(c, "invalid "+name)
	}
	return &value, nil
}

func QueryMatchingSelfOrID(c *fiber.Ctx, name string, expectedID int64, mismatchMessage string) (*int64, error) {
	value, err := QuerySelfOrID(c, name)
	if err != nil || value == nil {
		return value, err
	}
	if *value != expectedID {
		if mismatchMessage == "" {
			mismatchMessage = name + " must match authenticated user"
		}
		return nil, JSONError(c, fiber.StatusForbidden, mismatchMessage)
	}
	return value, nil
}

func RequireQueryMatchingSelfOrID(c *fiber.Ctx, name string, expectedID int64, requiredMessage string, mismatchMessage string) (int64, error) {
	value, err := QueryMatchingSelfOrID(c, name, expectedID, mismatchMessage)
	if err != nil {
		return 0, err
	}
	if value == nil {
		if requiredMessage == "" {
			requiredMessage = name + " query is required"
		}
		return 0, BadRequest(c, requiredMessage)
	}
	return *value, nil
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
	return ErrorResponse{Status: status, Body: ErrorMap(message)}
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

func NormalizePageSizeWithDefault(size int, defaultSize int) int {
	if defaultSize <= 0 {
		defaultSize = DefaultPageSize
	}
	if size <= 0 {
		return defaultSize
	}
	return ClampInt(size, MinPageSize, MaxPageSize)
}

func PageLimit(c *fiber.Ctx, defaultLimit int) (int, int) {
	return MaxIntQuery(c.QueryInt("page", DefaultPage), DefaultPage), ClampInt(c.QueryInt("limit", defaultLimit), MinPageSize, MaxPageSize)
}

func LimitOffset(c *fiber.Ctx, defaultLimit int) (int, int) {
	return ClampInt(c.QueryInt("limit", defaultLimit), MinPageSize, MaxPageSize), MaxInt(c.QueryInt("offset", 0), 0)
}

type QueryParser struct {
	c   *fiber.Ctx
	err error
}

func NewQueryParser(c *fiber.Ctx) *QueryParser {
	return &QueryParser{c: c}
}

func QueryParams(c *fiber.Ctx) *QueryParser {
	return NewQueryParser(c)
}

func (p *QueryParser) Err() error {
	return p.err
}

func (p *QueryParser) Page() int {
	if p == nil || p.c == nil {
		return DefaultPage
	}
	return MaxIntQuery(p.c.QueryInt("page", DefaultPage), DefaultPage)
}

func (p *QueryParser) String(name string) string {
	if p == nil || p.c == nil {
		return ""
	}
	return QueryString(p.c, name)
}

func (p *QueryParser) Int(name string, fallback int) int {
	if p == nil || p.c == nil {
		return fallback
	}
	return p.c.QueryInt(name, fallback)
}

func (p *QueryParser) Bool(name string, fallback bool) bool {
	if p == nil || p.c == nil {
		return fallback
	}
	return p.c.QueryBool(name, fallback)
}

func (p *QueryParser) OptionalPositiveInt64(name string) *int64 {
	if p.err != nil {
		return nil
	}
	value, err := ParseOptionalPositiveInt64Query(p.c, name)
	if err != nil {
		p.err = BadRequest(p.c, "invalid "+name)
		return nil
	}
	return value
}

func (p *QueryParser) Int64(name string) *int64 {
	return p.OptionalPositiveInt64(name)
}

func (p *QueryParser) RequiredPositiveInt64(name string) int64 {
	if p.err != nil {
		return 0
	}
	value, err := ParseRequiredPositiveInt64Query(p.c, name)
	if err != nil {
		p.err = BadRequest(p.c, err.Error())
		return 0
	}
	return value
}

func (p *QueryParser) OptionalDate(name string) *time.Time {
	if p.err != nil {
		return nil
	}
	value, err := ParseOptionalDateQuery(p.c, name)
	if err != nil {
		p.err = BadRequest(p.c, name+" must be YYYY-MM-DD")
		return nil
	}
	return value
}

func (p *QueryParser) PageLimit(defaultLimit int) (int, int) {
	if p == nil || p.c == nil {
		return DefaultPage, NormalizePageSizeWithDefault(0, defaultLimit)
	}
	return PageLimit(p.c, defaultLimit)
}

func (p *QueryParser) PageSize(defaultSize int) (int, int) {
	if p == nil || p.c == nil {
		return DefaultPage, NormalizePageSizeWithDefault(0, defaultSize)
	}
	return MaxIntQuery(p.c.QueryInt("page", DefaultPage), DefaultPage), NormalizePageSizeWithDefault(p.c.QueryInt("page_size", defaultSize), defaultSize)
}

func (p *QueryParser) PageSizeValue(defaultSize int) int {
	if p == nil || p.c == nil {
		return NormalizePageSizeWithDefault(0, defaultSize)
	}
	return NormalizePageSizeWithDefault(p.c.QueryInt("page_size", defaultSize), defaultSize)
}

// RequireInt64Param อ่าน path parameter เป็น int64 (ต้องมีค่า > 0)
func RequireInt64Param(c *fiber.Ctx, name string) (int64, error) {
	val, err := c.ParamsInt(name)
	if err != nil || val <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return int64(val), nil
}

func RequirePathID(c *fiber.Ctx, name string) (int64, error) {
	return RequireInt64Param(c, name)
}

// RequireStringParam อ่าน path parameter เป็น string (ต้องไม่เป็นค่าว่าง)
func RequireStringParam(c *fiber.Ctx, name string) (string, error) {
	val := ParamString(c, name)
	if val == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, name+" is required")
	}
	return val, nil
}

// OptionalInt64Param อ่าน path parameter เป็น int64 (ถ้าว่างกลับ nil)
func OptionalInt64Param(c *fiber.Ctx, name string) (*int64, error) {
	val := ParamString(c, name)
	if val == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil || parsed <= 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return &parsed, nil
}

// WriteAPIError เขียน standardized API error response
func WriteAPIError(c *fiber.Ctx, apiErr *APIError) error {
	if apiErr.Status == 0 {
		apiErr.Status = fiber.StatusInternalServerError
	}
	return c.Status(apiErr.Status).JSON(apiErr)
}

// NewAPIError สร้าง APIError ใหม่
func NewAPIError(status int, errorCode string, message string) *APIError {
	return &APIError{
		Status:    status,
		ErrorCode: errorCode,
		Message:   message,
	}
}

// NewAPIErrorWithDetails สร้าง APIError พร้อม details
func NewAPIErrorWithDetails(status int, errorCode string, message string, details map[string]interface{}) *APIError {
	return &APIError{
		Status:    status,
		ErrorCode: errorCode,
		Message:   message,
		Details:   details,
	}
}

// Common APIError creators
func BadRequestAPIError(code string, message string) *APIError {
	return NewAPIError(fiber.StatusBadRequest, code, message)
}

func NotFoundAPIError(code string, message string) *APIError {
	return NewAPIError(fiber.StatusNotFound, code, message)
}

func ConflictAPIError(code string, message string) *APIError {
	return NewAPIError(fiber.StatusConflict, code, message)
}

func ForbiddenAPIError(code string, message string) *APIError {
	return NewAPIError(fiber.StatusForbidden, code, message)
}

func UnauthorizedAPIError(code string, message string) *APIError {
	return NewAPIError(fiber.StatusUnauthorized, code, message)
}

func InternalServerAPIError(code string, message string) *APIError {
	return NewAPIError(fiber.StatusInternalServerError, code, message)
}

type ListResponse struct {
	Data    interface{} `json:"data"`
	Total   int         `json:"total"`
	Page    int         `json:"page,omitempty"`
	Limit   int         `json:"limit,omitempty"`
	HasMore bool        `json:"has_more,omitempty"`
}

type PaginatedListResponse struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteListResponse(c *fiber.Ctx, data interface{}, total int) error {
	return c.JSON(ListResponse{
		Data:  data,
		Total: total,
	})
}

func WriteSuccess(c *fiber.Ctx, message string, data interface{}) error {
	return c.JSON(SuccessResponse{
		Message: message,
		Data:    data,
	})
}

func CalculateOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * limit
}
