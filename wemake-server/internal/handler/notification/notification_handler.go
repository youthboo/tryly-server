package notification

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	notificationservice "github.com/yourusername/wemake/internal/service/notification"
)

type NotificationHandler struct {
	service *notificationservice.NotificationService
}

func NewNotificationHandler(service *notificationservice.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

func (h *NotificationHandler) List(c *fiber.Ctx) error {
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}
	query := helper.QueryParams(c)
	page, limit := query.PageLimit(helper.DefaultPageSize)
	unreadOnly := query.Bool("unread", false)
	items, total, unreadCount, err := h.service.ListPaginated(userID, page, limit, unreadOnly)
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch notifications")
	}
	return c.JSON(fiber.Map{
		"page":         page,
		"limit":        limit,
		"total":        total,
		"unread_count": unreadCount,
		"data":         items,
	})
}

func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}
	notiID, err := helper.RequireInt64Param(c, "noti_id")
	if err != nil {
		return err
	}
	if err := h.service.MarkAsRead(int64(notiID), userID); err != nil {
		return helper.JSONInternal(c, "failed to update notification")
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}
	count, err := h.service.GetUnreadCount(userID)
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch unread count")
	}
	return c.JSON(fiber.Map{"count": count})
}

func (h *NotificationHandler) MarkAllRead(c *fiber.Ctx) error {
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}
	updated, err := h.service.MarkAllRead(userID)
	if err != nil {
		return helper.JSONInternal(c, "failed to update notifications")
	}
	return c.JSON(fiber.Map{"updated": updated})
}

func (h *NotificationHandler) SoftDelete(c *fiber.Ctx) error {
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}
	notiID, err := helper.ParsePositiveInt64Param(c, "noti_id")
	if err != nil {
		return helper.BadRequestError(c, "invalid noti_id")
	}
	if err := h.service.SoftDelete(notiID, userID); err != nil {
		return helper.JSONInternal(c, "failed to delete notification")
	}
	return c.JSON(fiber.Map{"message": "notification deleted"})
}
