package handler

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/repository"
)

type AdminRFQHandler struct {
	repo  *repository.RFQRepository
	audit *repository.AdminAuditRepository
}

func NewAdminRFQHandler(repo *repository.RFQRepository, audit *repository.AdminAuditRepository) *AdminRFQHandler {
	return &AdminRFQHandler{repo: repo, audit: audit}
}

func (h *AdminRFQHandler) List(c *fiber.Ctx) error {
	filter := domain.AdminRFQFilter{
		Status:   strings.TrimSpace(c.Query("status")),
		Search:   strings.TrimSpace(c.Query("search")),
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("page_size", 20),
	}
	userID, err := parseOptionalPositiveInt64Query(c, "user_id")
	if err != nil {
		return badRequest(c, "invalid user_id")
	}
	filter.UserID = userID
	categoryID, err := parseOptionalPositiveInt64Query(c, "category_id")
	if err != nil {
		return badRequest(c, "invalid category_id")
	}
	filter.CategoryID = categoryID
	dateFrom, err := parseOptionalDateQuery(c, "date_from")
	if err != nil {
		return badRequest(c, "date_from must be YYYY-MM-DD")
	}
	filter.DateFrom = dateFrom
	dateTo, err := parseOptionalDateQuery(c, "date_to")
	if err != nil {
		return badRequest(c, "date_to must be YYYY-MM-DD")
	}
	filter.DateTo = dateTo
	items, total, err := h.repo.ListAdmin(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch rfqs"})
	}
	return c.JSON(fiber.Map{"data": items, "pagination": domain.Pagination{Page: maxInt(filter.Page, 1), PageSize: normalizePageSize(filter.PageSize), Total: total}})
}

func (h *AdminRFQHandler) GetByID(c *fiber.Ctx) error {
	rfqID, err := parsePositiveInt64Param(c, "rfq_id")
	if err != nil {
		return badRequest(c, "invalid rfq_id")
	}
	item, err := h.repo.GetAdminDetail(rfqID)
	if err != nil {
		return writeServiceError(c, err, "failed to fetch rfq", notFoundCase(errNotFound, "rfq not found"))
	}
	return c.JSON(item)
}

func (h *AdminRFQHandler) PatchStatus(c *fiber.Ctx) error {
	rfqID, err := parsePositiveInt64Param(c, "rfq_id")
	if err != nil {
		return badRequest(c, "invalid rfq_id")
	}
	var req struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := requireBody(c, &req); err != nil {
		return err
	}
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status != "CL" && status != "CC" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be CL or CC"})
	}
	if err := h.repo.UpdateStatusAdmin(rfqID, status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update rfq status"})
	}
	actorID, _ := getUserIDFromHeader(c)
	payload, _ := json.Marshal(map[string]interface{}{"status": status, "reason": req.Reason})
	ip := c.IP()
	_ = h.audit.Insert(&domain.AdminAuditLog{ActorID: actorID, Action: "RFQ_STATUS_CHANGE", TargetType: "rfq", TargetID: strconv.FormatInt(rfqID, 10), Payload: payload, IPAddress: &ip})
	return c.JSON(fiber.Map{"rfq_id": rfqID, "status": status})
}
