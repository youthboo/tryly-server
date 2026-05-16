package user

import (
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	userservice "github.com/yourusername/wemake/internal/service/user"
)

type CertificateHandler struct {
	service *userservice.CertificateService
}

func NewCertificateHandler(service *userservice.CertificateService) *CertificateHandler {
	return &CertificateHandler{service: service}
}

func (h *CertificateHandler) ListByFactory(c *fiber.Ctx) error {
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil {
		return badRequest(c, "invalid factory_id")
	}
	items, err := h.service.ListByFactoryID(int64(factoryID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch certificates"})
	}
	return c.JSON(items)
}

func (h *CertificateHandler) Create(c *fiber.Ctx) error {
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid factory_id"})
	}
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	// Assuming a factory user can only upload their own certificates
	if int64(factoryID) != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}

	var req domain.FactoryCertificate
	if err := requireBody(c, &req); err != nil {
		return err
	}
	req.FactoryID = int64(factoryID)

	if err := h.service.Create(&req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to upload certificate"})
	}
	return c.Status(fiber.StatusCreated).JSON(req)
}

func (h *CertificateHandler) Delete(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return unauthorized(c)
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || int64(factoryID) != userID {
		return jsonError(c, fiber.StatusForbidden, "forbidden")
	}
	mapID, err := parsePositiveInt64Param(c, "map_id")
	if err != nil {
		return badRequest(c, "invalid map_id")
	}
	if err := h.service.DeleteByMapID(int64(factoryID), mapID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if fallbackErr := h.service.DeleteByCertID(int64(factoryID), mapID); fallbackErr == nil {
				return c.SendStatus(fiber.StatusNoContent)
			}
			return jsonError(c, fiber.StatusNotFound, "certificate mapping not found")
		}
		return writeServiceError(c, err, "failed to delete certificate", notFoundCase(sql.ErrNoRows, "certificate mapping not found"))
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CertificateHandler) DeleteByCertID(c *fiber.Ctx) error {
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return unauthorized(c)
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || int64(factoryID) != userID {
		return jsonError(c, fiber.StatusForbidden, "forbidden")
	}
	certID, err := parsePositiveInt64Param(c, "cert_id")
	if err != nil {
		return badRequest(c, "invalid cert_id")
	}
	if err := h.service.DeleteByCertID(int64(factoryID), certID); err != nil {
		return writeServiceError(c, err, "failed to delete certificate", notFoundCase(sql.ErrNoRows, "certificate mapping not found"))
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CertificateHandler) PatchByCertID(c *fiber.Ctx) error {
	type reqBody struct {
		DocumentURL *string `json:"document_url"`
		ExpireDate  *string `json:"expire_date"`
		CertNumber  *string `json:"cert_number"`
	}
	userID, err := getUserIDFromHeader(c)
	if err != nil {
		return unauthorized(c)
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || int64(factoryID) != userID {
		return jsonError(c, fiber.StatusForbidden, "forbidden")
	}
	certID, err := parsePositiveInt64Param(c, "cert_id")
	if err != nil {
		return badRequest(c, "invalid cert_id")
	}
	var req reqBody
	if err := requireBody(c, &req); err != nil {
		return err
	}
	if req.DocumentURL == nil && req.ExpireDate == nil && req.CertNumber == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "at least one field is required"})
	}
	if err := h.service.PatchByCertID(int64(factoryID), certID, req.DocumentURL, req.ExpireDate, req.CertNumber); err != nil {
		return writeServiceError(c, err, "failed to update certificate", notFoundCase(sql.ErrNoRows, "certificate mapping not found"))
	}
	items, err := h.service.ListByFactoryID(int64(factoryID))
	if err != nil {
		return c.JSON(fiber.Map{"message": "certificate updated"})
	}
	for _, item := range items {
		if item.CertID == certID {
			return c.JSON(item)
		}
	}
	return c.JSON(fiber.Map{"message": "certificate updated"})
}
