package frontend

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	frontendrepo "github.com/yourusername/wemake/internal/repository/frontend"
)

type SessionHandler struct {
	repo *frontendrepo.SessionRepository
}

func NewSessionHandler(repo *frontendrepo.SessionRepository) *SessionHandler {
	return &SessionHandler{repo: repo}
}

func (h *SessionHandler) GetSession(c *fiber.Ctx) error {
	userID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}

	resp, err := h.repo.GetSession(userID)
	if err != nil {
		return helper.WriteServiceError(c, err, "failed to fetch session",
			helper.NotFoundCase(helper.ErrNotFound, "user not found"))
	}
	return c.JSON(resp)
}
