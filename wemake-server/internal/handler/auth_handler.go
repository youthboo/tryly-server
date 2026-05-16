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
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		case service.ErrInvalidRole, service.ErrMissingRoleData:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		default:
			logger.Error("user registration failed", "role", req.Role, "email", req.Email, "err", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "failed to register",
				"details": err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(result)
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
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		case service.ErrUserInactive:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to login"})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to process forgot password"})
	}

	resp := fiber.Map{
		"message": "if the account exists, reset instructions have been generated",
	}
	if token != "" {
		resp["reset_token"] = token
	}

	return c.JSON(resp)
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
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to reset password"})
	}

	return c.JSON(fiber.Map{"message": "password reset successful"})
}
