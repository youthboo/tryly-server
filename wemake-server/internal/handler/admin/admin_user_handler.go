package admin

import (
	"github.com/yourusername/wemake/internal/helper"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/repository"
	"github.com/yourusername/wemake/internal/service"
)

type AdminUserHandler struct {
	authService *service.AuthService
	authRepo    *repository.AuthRepository
}

func NewAdminUserHandler(authService *service.AuthService, authRepo *repository.AuthRepository) *AdminUserHandler {
	return &AdminUserHandler{authService: authService, authRepo: authRepo}
}

func (h *AdminUserHandler) Create(c *fiber.Ctx) error {
	actorID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	actor, err := h.authService.GetUserByID(actorID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req dto.CreateAdminUserRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	displayName := req.FirstName + " " + req.LastName
	item, err := h.authService.RegisterAdmin(service.RegisterAdminInput{
		Role:        strings.TrimSpace(req.Role),
		Email:       strings.TrimSpace(req.Email),
		Password:    req.Password,
		DisplayName: displayName,
		Department:  nil,
		CreatedBy:   &actorID,
	}, actor.Role)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(item.User)
}

func (h *AdminUserHandler) List(c *fiber.Ctx) error {
	items, err := h.authRepo.ListAdminUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch admin users"})
	}
	return c.JSON(fiber.Map{"data": items})
}
