package admin

import (
	"encoding/json"
	"github.com/yourusername/wemake/internal/helper"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/dto"
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
	filter := domain.AdminRFQFilter{
		Status:   strings.TrimSpace(c.Query("status")),
		Search:   strings.TrimSpace(c.Query("search")),
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("page_size", 20),
	}
	userID, err := helper.ParseOptionalPositiveInt64Query(c, "user_id")
	if err != nil {
		return helper.BadRequest(c, "invalid user_id")
	}
	filter.UserID = userID
	categoryID, err := helper.ParseOptionalPositiveInt64Query(c, "category_id")
	if err != nil {
		return helper.BadRequest(c, "invalid category_id")
	}
	filter.CategoryID = categoryID
	dateFrom, err := helper.ParseOptionalDateQuery(c, "date_from")
	if err != nil {
		return helper.BadRequest(c, "date_from must be YYYY-MM-DD")
	}
	filter.DateFrom = dateFrom
	dateTo, err := helper.ParseOptionalDateQuery(c, "date_to")
	if err != nil {
		return helper.BadRequest(c, "date_to must be YYYY-MM-DD")
	}
	filter.DateTo = dateTo
	items, total, err := h.repo.ListAdmin(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch rfqs"})
	}
	return c.JSON(fiber.Map{"data": items, "pagination": domain.Pagination{Page: helper.MaxInt(filter.Page, 1), PageSize: helper.NormalizePageSize(filter.PageSize), Total: total}})
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
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status != "CL" && status != "CC" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be CL or CC"})
	}
	if err := h.repo.UpdateStatusAdmin(rfqID, status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update rfq status"})
	}
	actorID, _ := helper.UserIDFromHeader(c)
	notes := ""
	if req.Notes != nil {
		notes = *req.Notes
	}
	payload, _ := json.Marshal(map[string]interface{}{"status": status, "notes": notes})
	ip := c.IP()
	_ = h.audit.Insert(&domain.AdminAuditLog{ActorID: actorID, Action: "RFQ_STATUS_CHANGE", TargetType: "rfq", TargetID: strconv.FormatInt(rfqID, 10), Payload: payload, IPAddress: &ip})
	return c.JSON(fiber.Map{"rfq_id": rfqID, "status": status})
}
