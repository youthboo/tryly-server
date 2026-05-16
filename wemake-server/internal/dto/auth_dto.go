package dto

// Auth Request DTOs
type RegisterRequest struct {
	Role          string `json:"role" validate:"notblank"`
	Email         string `json:"email" validate:"notblank"`
	Phone         string `json:"phone"`
	Password      string `json:"password" validate:"notblank"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	FactoryName   string `json:"factory_name"`
	FactoryTypeID int64  `json:"factory_type_id"`
	TaxID         string `json:"tax_id"`
	ProvinceID    *int64 `json:"province_id"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"notblank"`
	Password string `json:"password" validate:"notblank"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"notblank"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"notblank"`
	NewPassword string `json:"new_password" validate:"notblank"`
}
