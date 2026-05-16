package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	userservice "github.com/yourusername/wemake/internal/service/user"
)

type FavoriteHandler struct {
	service *userservice.FavoriteService
}

func NewFavoriteHandler(service *userservice.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{service: service}
}

func (h *FavoriteHandler) List(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return unauthorized(c)
	}
	items, err := h.service.ListByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch favorites"})
	}
	return c.JSON(items)
}

func (h *FavoriteHandler) Add(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return unauthorized(c)
	}

	var req domain.Favorite
	if err := requireBody(c, &req); err != nil {
		return err
	}
	req.UserID = userID

	if err := h.service.Add(&req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to add favorite"})
	}
	return c.Status(fiber.StatusCreated).JSON(req)
}

func (h *FavoriteHandler) Remove(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return unauthorized(c)
	}
	showcaseID, err := c.ParamsInt("showcase_id")
	if err != nil {
		return badRequest(c, "invalid showcase_id")
	}
	if err := h.service.Remove(userID, int64(showcaseID)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to remove favorite"})
	}
	return c.JSON(fiber.Map{"success": true})
}
