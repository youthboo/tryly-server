package quotation

import (
	"database/sql"
	"errors"
	"github.com/yourusername/wemake/internal/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/dto"
	quotationservice "github.com/yourusername/wemake/internal/service/quotation"
)

type QuotationTemplateHandler struct {
	service *quotationservice.QuotationTemplateService
}

func NewQuotationTemplateHandler(svc *quotationservice.QuotationTemplateService) *QuotationTemplateHandler {
	return &QuotationTemplateHandler{service: svc}
}

// GET /quotation-templates
func (h *QuotationTemplateHandler) List(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
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
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req dto.CreateQuotationTemplateRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	tmpl := &domain.QuotationTemplate{
		FactoryID: userID,
		TemplateName: req.Name,
		Note: req.Description,
	}
	if req.IsActive != nil {
		tmpl.IsActive = *req.IsActive
	}
	if err := h.service.Create(tmpl); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create quotation template"})
	}
	return c.Status(fiber.StatusCreated).JSON(tmpl)
}

// PATCH /quotation-templates/:template_id
func (h *QuotationTemplateHandler) Patch(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	templateID, err := helper.ParsePositiveInt64Param(c, "template_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid template_id"})
	}
	var req dto.PatchQuotationTemplateRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	tmpl := &domain.QuotationTemplate{
		TemplateID: templateID,
		FactoryID: userID,
	}
	if req.Name != nil {
		tmpl.TemplateName = *req.Name
	}
	if req.Description != nil {
		tmpl.Note = req.Description
	}
	if req.IsActive != nil {
		tmpl.IsActive = *req.IsActive
	}
	if err := h.service.Update(tmpl); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "quotation template not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update quotation template"})
	}
	return c.JSON(tmpl)
}

// DELETE /quotation-templates/:template_id
func (h *QuotationTemplateHandler) Delete(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	templateID, err := helper.ParsePositiveInt64Param(c, "template_id")
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
