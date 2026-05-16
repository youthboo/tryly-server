package handler

import (
	"errors"

	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/service"
)

type PlatformConfigHandler struct {
	service *service.PlatformConfigService
	auth    *service.AuthService
}

func NewPlatformConfigHandler(service *service.PlatformConfigService, auth *service.AuthService) *PlatformConfigHandler {
	return &PlatformConfigHandler{service: service, auth: auth}
}

func (h *PlatformConfigHandler) GetActive(c *fiber.Ctx) error {
	item, err := h.service.GetActive()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch platform config"})
	}
	return c.JSON(item)
}

func (h *PlatformConfigHandler) ListHistory(c *fiber.Ctx) error {
	items, err := h.service.ListHistory()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch platform config history"})
	}
	return c.JSON(items)
}

func (h *PlatformConfigHandler) ListAll(c *fiber.Ctx) error {
	items, err := h.service.ListAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch platform configs"})
	}
	return c.JSON(fiber.Map{"configs": items, "total": len(items)})
}

func (h *PlatformConfigHandler) Create(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req dto.CreateConfigVersionRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	promoStartAt, err := helper.ParseOptionalRFC3339Value(req.PromoStartAt, "promo_start_at")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "promo_start_at must be RFC3339"})
	}
	promoEndAt, err := helper.ParseOptionalRFC3339Value(req.PromoEndAt, "promo_end_at")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "promo_end_at must be RFC3339"})
	}
	cfg := &domain.PlatformConfig{
		DefaultCommissionRate: req.DefaultCommissionRate,
		PromoCommissionRate:   req.PromoCommissionRate,
		PromoStartAt:          promoStartAt,
		PromoEndAt:            promoEndAt,
		PromoLabel:            req.PromoLabel,
		VatRate:               req.VatRate,
		CurrencyCode:          req.CurrencyCode,
		CreatedBy:             &userID,
	}
	if err := h.service.CreateVersion(cfg); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create platform config"})
	}
	return c.Status(fiber.StatusCreated).JSON(cfg)
}

func (h *PlatformConfigHandler) CreateConfig(c *fiber.Ctx) error {
	actorID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req domain.CreatePlatformConfigRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	ip := c.IP()
	item, err := h.service.CreateConfig(req, actorID, &ip)
	if err != nil {
		if errors.Is(err, service.ErrPlatformConfigValidation) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid platform config payload"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create platform config"})
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *PlatformConfigHandler) UpdateConfig(c *fiber.Ctx) error {
	configID, err := helper.ParsePositiveInt64Param(c, "config_id")
	if err != nil || configID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid config_id"})
	}
	actorID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req domain.UpdatePlatformConfigRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	ip := c.IP()
	item, err := h.service.UpdateConfig(configID, req, actorID, &ip)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPlatformConfigValidation):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid platform config payload"})
		case errors.Is(err, service.ErrPlatformConfigNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "platform config not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update platform config"})
		}
	}
	return c.JSON(item)
}

func (h *PlatformConfigHandler) DeleteConfig(c *fiber.Ctx) error {
	configID, err := helper.ParsePositiveInt64Param(c, "config_id")
	if err != nil || configID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid config_id"})
	}
	actorID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	ip := c.IP()
	err = h.service.DeleteConfig(configID, actorID, &ip)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPlatformConfigValidation), errors.Is(err, service.ErrPlatformDefaultDelete):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, service.ErrPlatformConfigNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "platform config not found"})
		case errors.Is(err, service.ErrPlatformConfigInUse):
			count, _ := h.service.FactoriesUsingConfig(configID)
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "ไม่สามารถลบได้ มีโรงงาน " + strconv.Itoa(count) + " แห่งกำลังใช้ config นี้อยู่"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete platform config"})
		}
	}
	return c.SendStatus(fiber.StatusNoContent)
}
