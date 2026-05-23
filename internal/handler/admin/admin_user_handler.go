package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	authrepo "github.com/yourusername/wemake/internal/repository/auth"
	authservice "github.com/yourusername/wemake/internal/service/auth"
)

type AdminUserHandler struct {
	authService *authservice.AuthService
	authRepo    *authrepo.AuthRepository
}

func NewAdminUserHandler(authService *authservice.AuthService, authRepo *authrepo.AuthRepository) *AdminUserHandler {
	return &AdminUserHandler{authService: authService, authRepo: authRepo}
}

func (h *AdminUserHandler) Create(c *fiber.Ctx) error {
	actorID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}
	actor, err := h.authService.GetUserByID(actorID)
	if err != nil {
		return helper.UnauthorizedError(c, "unauthorized")
	}
	var req dto.CreateAdminUserRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	displayName := req.FirstName + " " + req.LastName
	item, err := h.authService.RegisterAdmin(authservice.RegisterAdminInput{
		Role:        helper.DereferenceString(&req.Role, ""),
		Email:       helper.DereferenceString(&req.Email, ""),
		Password:    req.Password,
		DisplayName: displayName,
		Department:  nil,
		CreatedBy:   &actorID,
	}, actor.Role)
	if err != nil {
		return helper.BadRequestError(c, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(item.User)
}

func (h *AdminUserHandler) List(c *fiber.Ctx) error {
	items, err := h.authRepo.ListAdminUsers()
	if err != nil {
		return helper.InternalServerError(c, "failed to fetch admin users")
	}
	return c.JSON(fiber.Map{"data": items})
}
