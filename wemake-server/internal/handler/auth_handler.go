package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	"github.com/yourusername/wemake/internal/logger"
	"github.com/yourusername/wemake/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
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

	result, err := h.service.Register(service.RegisterInput{
		Role:          req.Role,
		Email:         req.Email,
		Phone:         req.Phone,
		Password:      req.Password,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		FactoryName:   req.FactoryName,
		FactoryTypeID: req.FactoryTypeID,
		TaxID:         req.TaxID,
		ProvinceID:    req.ProvinceID,
	})
	if err != nil {
		switch err {
		case service.ErrEmailAlreadyExists:
			return helper.WriteAPIError(c, helper.ConflictAPIError("EMAIL_EXISTS", "email already exists"))
		case service.ErrInvalidRole, service.ErrMissingRoleData:
			return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_ROLE", err.Error()))
		default:
			logger.Error("user registration failed", "role", req.Role, "email", req.Email, "err", err)
			return helper.WriteAPIError(c, helper.InternalServerError("REGISTER_FAILED", "failed to register"))
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
		case service.ErrInvalidCredentials:
			return helper.WriteAPIError(c, helper.UnauthorizedAPIError("INVALID_CREDENTIALS", "invalid email or password"))
		case service.ErrUserInactive:
			return helper.WriteAPIError(c, helper.ForbiddenAPIError("USER_INACTIVE", "account is inactive"))
		default:
			return helper.WriteAPIError(c, helper.InternalServerError("LOGIN_FAILED", "failed to login"))
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
		return helper.WriteAPIError(c, helper.InternalServerError("FORGOT_PASSWORD_FAILED", "failed to process forgot password"))
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
		if err == service.ErrInvalidResetToken {
			return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_RESET_TOKEN", "invalid or expired reset token"))
		}
		return helper.WriteAPIError(c, helper.InternalServerError("RESET_PASSWORD_FAILED", "failed to reset password"))
	}

	return helper.WriteSuccess(c, "password reset successful", nil)
}
