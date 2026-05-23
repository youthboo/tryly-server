package user

import (
	"database/sql"
	"errors"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"

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

var certificateNotFoundErrorMap = map[error]helper.ErrorResponse{
	sql.ErrNoRows: helper.ErrorMessage(fiber.StatusNotFound, "certificate mapping not found"),
}

func (h *CertificateHandler) ListByFactory(c *fiber.Ctx) error {
	factoryID, err := helper.RequireInt64Param(c, "factory_id")
	if err != nil {
		return err
	}
	items, err := h.service.ListByFactoryID(int64(factoryID))
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch certificates")
	}
	return c.JSON(items)
}

func (h *CertificateHandler) Create(c *fiber.Ctx) error {
	factoryID, err := helper.RequireInt64Param(c, "factory_id")
	if err != nil {
		return err
	}
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}

	// Assuming a factory user can only upload their own certificates
	if int64(factoryID) != userID {
		return helper.ForbiddenError(c, "forbidden")
	}

	var req domain.FactoryCertificate
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	req.FactoryID = int64(factoryID)

	if err := h.service.Create(&req); err != nil {
		return helper.JSONInternal(c, "failed to upload certificate")
	}
	return c.Status(fiber.StatusCreated).JSON(req)
}

func (h *CertificateHandler) Delete(c *fiber.Ctx) error {
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}
	factoryID, err := helper.RequireInt64Param(c, "factory_id")
	if err != nil || int64(factoryID) != userID {
		return helper.JSONError(c, fiber.StatusForbidden, "forbidden")
	}
	mapID, err := helper.ParsePositiveInt64Param(c, "map_id")
	if err != nil {
		return helper.BadRequest(c, "invalid map_id")
	}
	if err := h.service.DeleteByMapID(int64(factoryID), mapID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if fallbackErr := h.service.DeleteByCertID(int64(factoryID), mapID); fallbackErr == nil {
				return c.SendStatus(fiber.StatusNoContent)
			}
			return helper.MapServiceError(c, sql.ErrNoRows, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to delete certificate"), certificateNotFoundErrorMap)
		}
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to delete certificate"), certificateNotFoundErrorMap)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CertificateHandler) DeleteByCertID(c *fiber.Ctx) error {
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}
	factoryID, err := helper.RequireInt64Param(c, "factory_id")
	if err != nil || int64(factoryID) != userID {
		return helper.JSONError(c, fiber.StatusForbidden, "forbidden")
	}
	certID, err := helper.ParsePositiveInt64Param(c, "cert_id")
	if err != nil {
		return helper.BadRequest(c, "invalid cert_id")
	}
	if err := h.service.DeleteByCertID(int64(factoryID), certID); err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to delete certificate"), certificateNotFoundErrorMap)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CertificateHandler) PatchByCertID(c *fiber.Ctx) error {
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}
	factoryID, err := helper.RequireInt64Param(c, "factory_id")
	if err != nil || int64(factoryID) != userID {
		return helper.JSONError(c, fiber.StatusForbidden, "forbidden")
	}
	certID, err := helper.ParsePositiveInt64Param(c, "cert_id")
	if err != nil {
		return helper.BadRequest(c, "invalid cert_id")
	}
	var req dto.PatchCertificateRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	v := domain.NewValidationCollector()
	v.AddIf(req.DocumentURL == nil && req.ExpireDate == nil && req.CertNumber == nil, "fields", "at least one field is required")
	if err := helper.ValidateRequest(c, v); err != nil {
		return err
	}
	if err := h.service.PatchByCertID(int64(factoryID), certID, req.DocumentURL, req.ExpireDate, req.CertNumber); err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to update certificate"), certificateNotFoundErrorMap)
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
