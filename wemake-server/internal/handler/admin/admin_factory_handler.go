package admin

import (
	"errors"
	"github.com/yourusername/wemake/internal/helper"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/dto"
	adminrepo "github.com/yourusername/wemake/internal/repository/admin"
	adminservice "github.com/yourusername/wemake/internal/service/admin"
)

type AdminFactoryHandler struct {
	repo    *adminrepo.AdminFactoryRepository
	service *adminservice.AdminFactoryService
}

func NewAdminFactoryHandler(repo *adminrepo.AdminFactoryRepository, service *adminservice.AdminFactoryService) *AdminFactoryHandler {
	return &AdminFactoryHandler{repo: repo, service: service}
}

func (h *AdminFactoryHandler) List(c *fiber.Ctx) error {
	filter := domain.AdminFactoryFilter{
		ApprovalStatus: strings.TrimSpace(c.Query("approval_status")),
		Search:         strings.TrimSpace(c.Query("search")),
		Page:           c.QueryInt("page", 1),
		PageSize:       c.QueryInt("page_size", 20),
	}
	if v := strings.TrimSpace(c.Query("is_verified")); v != "" {
		isVerified := strings.EqualFold(v, "true") || v == "1"
		filter.IsVerified = &isVerified
	}
	items, total, err := h.repo.ListAdmin(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch factories"})
	}
	return c.JSON(fiber.Map{"data": items, "pagination": domain.Pagination{Page: helper.MaxInt(filter.Page, 1), PageSize: helper.NormalizePageSize(filter.PageSize), Total: total}})
}

func (h *AdminFactoryHandler) GetByID(c *fiber.Ctx) error {
	factoryID, err := helper.ParsePositiveInt64Param(c, "factory_id")
	if err != nil {
		return helper.BadRequest(c, "invalid factory_id")
	}
	item, err := h.service.HydrateAdminDetail(factoryID)
	if err != nil {
		return helper.WriteServiceError(c, err, "failed to fetch factory", helper.NotFoundCase(helper.ErrNotFound, "factory not found"))
	}
	return c.JSON(item)
}

func (h *AdminFactoryHandler) Approve(c *fiber.Ctx) error {
	return h.mutateFactoryState(c, h.service.Approve)
}

func (h *AdminFactoryHandler) Reject(c *fiber.Ctx) error {
	return h.mutateFactoryReasonState(c, h.service.Reject)
}

func (h *AdminFactoryHandler) Suspend(c *fiber.Ctx) error {
	return h.mutateFactoryReasonState(c, h.service.Suspend)
}

func (h *AdminFactoryHandler) Unsuspend(c *fiber.Ctx) error {
	return h.mutateFactoryState(c, h.service.Unsuspend)
}

func (h *AdminFactoryHandler) PatchVerification(c *fiber.Ctx) error {
	factoryID, actorID, err := parseFactoryActor(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	var req dto.PatchFactoryVerificationRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	ip := c.IP()
	isVerified := req.TaxIDVerified != nil && *req.TaxIDVerified
	note := ""
	if req.Notes != nil {
		note = *req.Notes
	}
	if err := h.service.ToggleVerification(factoryID, actorID, isVerified, note, &ip); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	item, _ := h.service.HydrateAdminDetail(factoryID)
	return c.JSON(fiber.Map{
		"factory_id":      factoryID,
		"is_verified":     item.IsVerified,
		"verified_at":     item.VerifiedAt,
		"verified_by":     item.VerifiedBy,
		"approval_status": item.ApprovalStatus,
	})
}

func (h *AdminFactoryHandler) GetFactoryConfig(c *fiber.Ctx) error {
	factoryID, err := helper.ParsePositiveInt64Param(c, "factory_id")
	if err != nil {
		return helper.BadRequest(c, "invalid factory_id")
	}
	item, err := h.service.GetFactoryConfig(factoryID)
	if err != nil {
		return helper.WriteServiceError(c, err, "failed to fetch factory config", helper.NotFoundCase(adminservice.ErrFactoryNotFound, "factory not found"))
	}
	return c.JSON(item)
}

func (h *AdminFactoryHandler) AssignFactoryConfig(c *fiber.Ctx) error {
	factoryID, actorID, err := parseFactoryActor(c)
	if err != nil {
		status := fiber.StatusBadRequest
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) && fiberErr.Code == fiber.StatusUnauthorized {
			status = fiber.StatusUnauthorized
		}
		return c.Status(status).JSON(fiber.Map{"error": err.Error()})
	}
	var req domain.AssignFactoryConfigRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	ip := c.IP()
	item, err := h.service.AssignFactoryConfig(factoryID, actorID, req, &ip)
	if err != nil {
		switch {
		case errors.Is(err, adminservice.ErrFactoryNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "factory not found"})
		case errors.Is(err, adminservice.ErrFactoryConfigMissing):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "platform config not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to assign factory config"})
		}
	}
	return c.JSON(item)
}

func (h *AdminFactoryHandler) mutateFactoryState(c *fiber.Ctx, fn func(int64, int64, string, *string) error) error {
	factoryID, actorID, err := parseFactoryActor(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	var req dto.ApproveFactoryRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	ip := c.IP()
	note := ""
	if req.Notes != nil {
		note = *req.Notes
	}
	if err := fn(factoryID, actorID, note, &ip); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	item, _ := h.service.HydrateAdminDetail(factoryID)
	return c.JSON(fiber.Map{"factory_id": factoryID, "approval_status": item.ApprovalStatus, "is_verified": item.IsVerified, "verified_at": item.VerifiedAt, "verified_by": item.VerifiedBy})
}

func (h *AdminFactoryHandler) mutateFactoryReasonState(c *fiber.Ctx, fn func(int64, int64, string, *string) error) error {
	factoryID, actorID, err := parseFactoryActor(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	var req dto.RejectFactoryRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	ip := c.IP()
	if err := fn(factoryID, actorID, req.Reason, &ip); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	item, _ := h.service.HydrateAdminDetail(factoryID)
	return c.JSON(fiber.Map{"factory_id": factoryID, "approval_status": item.ApprovalStatus, "is_verified": item.IsVerified, "rejection_reason": item.RejectionReason})
}

func parseFactoryActor(c *fiber.Ctx) (int64, int64, error) {
	factoryID, err := helper.ParsePositiveInt64Param(c, "factory_id")
	if err != nil || factoryID <= 0 {
		return 0, 0, fiber.NewError(fiber.StatusBadRequest, "invalid factory_id")
	}
	actorID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return 0, 0, fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	return factoryID, actorID, nil
}
