package me

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	meservice "github.com/yourusername/wemake/internal/service/me"
)

type MeRFQOrdersHandler struct {
	service *meservice.RFQOrdersService
}

func NewMeRFQOrdersHandler(service *meservice.RFQOrdersService) *MeRFQOrdersHandler {
	return &MeRFQOrdersHandler{service: service}
}

func (h *MeRFQOrdersHandler) ListRFQOrders(c *fiber.Ctx) error {
	userID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}

	result, err := h.service.List(userID)
	if err != nil {
		return helper.JSONError(c, fiber.StatusInternalServerError, "failed to list rfq-orders")
	}
	return c.JSON(result)
}

func (h *MeRFQOrdersHandler) GetRFQOrderDetail(c *fiber.Ctx) error {
	userID, err := helper.RequireUserID(c)
	if err != nil {
		return err
	}
	rfqID, err := helper.RequireInt64Param(c, "rfq_id")
	if err != nil {
		return err
	}

	detail, err := h.service.GetDetail(userID, rfqID)
	if err != nil {
		if errors.Is(err, meservice.ErrRFQOrderNotFound) {
			return helper.JSONError(c, fiber.StatusNotFound, "rfq not found")
		}
		return helper.JSONError(c, fiber.StatusInternalServerError, "failed to fetch rfq-orders detail")
	}
	return c.JSON(detail)
}
