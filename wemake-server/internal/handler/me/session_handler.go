package me

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	frontendservice "github.com/yourusername/wemake/internal/service/frontend"
)

type SessionHandler struct {
	frontend *frontendservice.FrontendService
}

func NewSessionHandler(frontend *frontendservice.FrontendService) *SessionHandler {
	return &SessionHandler{frontend: frontend}
}

func (h *SessionHandler) GetSession(c *fiber.Ctx) error {
	userID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}

	resp, err := h.frontend.GetSession(userID)
	if err != nil {
		return helper.WriteServiceError(c, err, "failed to fetch session",
			helper.NotFoundCase(helper.ErrNotFound, "user not found"))
	}
	return c.JSON(resp)
}
