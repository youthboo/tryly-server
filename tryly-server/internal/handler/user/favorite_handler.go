package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/helper"
	userservice "github.com/yourusername/wemake/internal/service/user"
)

type FavoriteHandler struct {
	service *userservice.FavoriteService
}

func NewFavoriteHandler(service *userservice.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{service: service}
}

func (h *FavoriteHandler) List(c *fiber.Ctx) error {
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}
	items, err := h.service.ListByUserID(userID)
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch favorites")
	}
	return c.JSON(items)
}

func (h *FavoriteHandler) Add(c *fiber.Ctx) error {
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}

	var req domain.Favorite
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	req.UserID = userID

	if err := h.service.Add(&req); err != nil {
		return helper.JSONInternal(c, "failed to add favorite")
	}
	return c.Status(fiber.StatusCreated).JSON(req)
}

func (h *FavoriteHandler) Remove(c *fiber.Ctx) error {
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}
	showcaseID, err := helper.RequireInt64Param(c, "showcase_id")
	if err != nil {
		return err
	}
	if err := h.service.Remove(userID, int64(showcaseID)); err != nil {
		return helper.JSONInternal(c, "failed to remove favorite")
	}
	return c.JSON(fiber.Map{"success": true})
}
