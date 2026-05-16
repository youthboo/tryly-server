package admin

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/domainutil"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	adminrepo "github.com/yourusername/wemake/internal/repository/admin"
)

type AdminRFQHandler struct {
	repo  *adminrepo.AdminRFQRepository
	audit *adminrepo.AdminAuditRepository
}

func NewAdminRFQHandler(repo *adminrepo.AdminRFQRepository, audit *adminrepo.AdminAuditRepository) *AdminRFQHandler {
	return &AdminRFQHandler{repo: repo, audit: audit}
}

func (h *AdminRFQHandler) List(c *fiber.Ctx) error {
	query := helper.QueryParams(c)
	page, pageSize := query.PageSize(helper.DefaultPageSize)
	filter := domain.AdminRFQFilter{
		Status:     query.String("status"),
		Search:     query.String("search"),
		Page:       page,
		PageSize:   pageSize,
		UserID:     query.OptionalPositiveInt64("user_id"),
		CategoryID: query.OptionalPositiveInt64("category_id"),
		DateFrom:   query.OptionalDate("date_from"),
		DateTo:     query.OptionalDate("date_to"),
	}
	if err := query.Err(); err != nil {
		return err
	}
	items, total, err := h.repo.ListAdmin(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch rfqs"})
	}
	return c.JSON(fiber.Map{"data": items, "pagination": domain.Pagination{Page: filter.Page, PageSize: filter.PageSize, Total: total}})
}

func (h *AdminRFQHandler) GetByID(c *fiber.Ctx) error {
	rfqID, err := helper.ParsePositiveInt64Param(c, "rfq_id")
	if err != nil {
		return helper.BadRequest(c, "invalid rfq_id")
	}
	item, err := h.repo.GetAdminDetail(rfqID)
	if err != nil {
		return helper.WriteServiceError(c, err, "failed to fetch rfq", helper.NotFoundCase(helper.ErrNotFound, "rfq not found"))
	}
	return c.JSON(item)
}

func (h *AdminRFQHandler) PatchStatus(c *fiber.Ctx) error {
	rfqID, err := helper.ParsePositiveInt64Param(c, "rfq_id")
	if err != nil {
		return helper.BadRequest(c, "invalid rfq_id")
	}
	var req dto.PatchRFQStatusRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	status := domainutil.NormalizeStatus(req.Status)
	if status != "CL" && status != "CC" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be CL or CC"})
	}
	if err := h.repo.UpdateStatusAdmin(rfqID, status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update rfq status"})
	}
	actorID := helper.OptionalActorID(c)
	notes := helper.DereferenceString(req.Notes, "")
	payload, _ := json.Marshal(map[string]interface{}{"status": status, "notes": notes})
	ip := c.IP()
	_ = h.audit.Insert(&domain.AdminAuditLog{ActorID: actorID, Action: "RFQ_STATUS_CHANGE", TargetType: "rfq", TargetID: strconv.FormatInt(rfqID, 10), Payload: payload, IPAddress: &ip})
	return c.JSON(fiber.Map{"rfq_id": rfqID, "status": status})
}
