package handler

import (
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/service"
)

type QuotationTemplateHandler struct {
	service *service.QuotationTemplateService
}

func NewQuotationTemplateHandler(svc *service.QuotationTemplateService) *QuotationTemplateHandler {
	return &QuotationTemplateHandler{service: svc}
}

// GET /quotation-templates
func (h *QuotationTemplateHandler) List(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	items, err := h.service.ListByFactoryID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch quotation templates"})
	}
	return c.JSON(items)
}

// POST /quotation-templates
func (h *QuotationTemplateHandler) Create(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req domain.QuotationTemplate
	if err := parseAndValidateBody(c, &req, map[string]string{
		"TemplateName": "template_name is required",
	}); err != nil {
		return err
	}
	req.FactoryID = userID
	if err := h.service.Create(&req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create quotation template"})
	}
	return c.Status(fiber.StatusCreated).JSON(req)
}

// PATCH /quotation-templates/:template_id
func (h *QuotationTemplateHandler) Patch(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	templateID, err := parsePositiveInt64Param(c, "template_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid template_id"})
	}
	var req domain.QuotationTemplate
	if err := requireBody(c, &req); err != nil {
		return err
	}
	req.TemplateID = templateID
	req.FactoryID = userID
	if err := h.service.Update(&req); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "quotation template not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update quotation template"})
	}
	return c.JSON(req)
}

// DELETE /quotation-templates/:template_id
func (h *QuotationTemplateHandler) Delete(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	templateID, err := parsePositiveInt64Param(c, "template_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid template_id"})
	}
	if err := h.service.Delete(templateID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "quotation template not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete quotation template"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
