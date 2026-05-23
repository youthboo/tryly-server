package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	"github.com/yourusername/wemake/internal/logger"
	authservice "github.com/yourusername/wemake/internal/service/auth"
)

type AuthHandler struct {
	service *authservice.AuthService
}

func NewAuthHandler(service *authservice.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := helper.ParseAndValidateBody(c, &req, map[string]string{
		"Role":     "role, email, and password are required",
		"Email":    "role, email, and password are required",
		"Password": "role, email, and password are required",
	}); err != nil {
		return err
	}

	result, err := h.service.Register(authservice.RegisterInput{
		Role:           req.Role,
		Email:          req.Email,
		Phone:          req.Phone,
		Password:       req.Password,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		FactoryName:    req.FactoryName,
		FactoryTypeID:  req.FactoryTypeID,
		TaxID:          req.TaxID,
		ProvinceID:     req.ProvinceID,
		CategoryIDs:    req.CategoryIDs,
		SubCategoryIDs: req.SubCategoryIDs,
		CertID:         req.CertID,
		DocumentURL:    req.DocumentURL,
		CertNumber:     req.CertNumber,
		CertExpireDate: req.CertExpireDate,
	})
	if err != nil {
		switch err {
		case authservice.ErrEmailAlreadyExists:
			return helper.WriteAPIError(c, helper.ConflictAPIError("EMAIL_EXISTS", "email already exists"))
		case authservice.ErrInvalidRole, authservice.ErrMissingRoleData:
			return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_ROLE", err.Error()))
		default:
			logger.Error("user registration failed", "role", req.Role, "email", req.Email, "err", err)
			return helper.WriteAPIError(c, helper.InternalServerAPIError("REGISTER_FAILED", "failed to register"))
		}
	}

	c.Status(fiber.StatusCreated)
	return c.JSON(result)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := helper.ParseAndValidateBody(c, &req, map[string]string{
		"Email":    "email and password are required",
		"Password": "email and password are required",
	}); err != nil {
		return err
	}

	result, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		switch err {
		case authservice.ErrInvalidCredentials:
			return helper.WriteAPIError(c, helper.UnauthorizedAPIError("INVALID_CREDENTIALS", "invalid email or password"))
		case authservice.ErrUserInactive:
			return helper.WriteAPIError(c, helper.ForbiddenAPIError("USER_INACTIVE", "account is inactive"))
		default:
			return helper.WriteAPIError(c, helper.InternalServerAPIError("LOGIN_FAILED", "failed to login"))
		}
	}

	return c.JSON(result)
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := helper.ParseAndValidateBody(c, &req, map[string]string{
		"Email": "email is required",
	}); err != nil {
		return err
	}

	token, err := h.service.ForgotPassword(req.Email)
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerAPIError("FORGOT_PASSWORD_FAILED", "failed to process forgot password"))
	}

	data := fiber.Map{}
	if token != "" {
		data["reset_token"] = token
	}

	return helper.WriteSuccess(c, "if the account exists, reset instructions have been generated", data)
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := helper.ParseAndValidateBody(c, &req, map[string]string{
		"Token":       "token and new_password are required",
		"NewPassword": "token and new_password are required",
	}); err != nil {
		return err
	}

	if err := h.service.ResetPassword(req.Token, req.NewPassword); err != nil {
		if err == authservice.ErrInvalidResetToken {
			return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_RESET_TOKEN", "invalid or expired reset token"))
		}
		return helper.WriteAPIError(c, helper.InternalServerAPIError("RESET_PASSWORD_FAILED", "failed to reset password"))
	}

	return helper.WriteSuccess(c, "password reset successful", nil)
}
