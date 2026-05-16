package admin

import (
	"encoding/json"
	"github.com/yourusername/wemake/internal/helper"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/dto"
	adminrepo "github.com/yourusername/wemake/internal/repository/admin"
	walletrepo "github.com/yourusername/wemake/internal/repository/wallet"
)

type AdminConfigHandler struct {
	commission *walletrepo.CommissionRepository
	audit      *adminrepo.AdminAuditRepository
}

func NewAdminConfigHandler(commission *walletrepo.CommissionRepository, audit *adminrepo.AdminAuditRepository) *AdminConfigHandler {
	return &AdminConfigHandler{commission: commission, audit: audit}
}

func (h *AdminConfigHandler) ListRules(c *fiber.Ctx) error {
	factoryID, err := helper.ParseOptionalPositiveInt64Query(c, "factory_id")
	if err != nil {
		return helper.BadRequest(c, "invalid factory_id")
	}
	items, err := h.commission.ListRules(factoryID, !strings.EqualFold(c.Query("active_only", "true"), "false"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch commission rules"})
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *AdminConfigHandler) CreateRule(c *fiber.Ctx) error {
	var req dto.CreateCommissionRuleRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	if req.CommissionRate < 0 || req.CommissionRate > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid commission rule payload"})
	}
	from := time.Now().UTC()
	actorID, _ := helper.UserIDFromHeader(c)
	item := &domain.CommissionRule{RatePercent: req.CommissionRate, EffectiveFrom: from, Note: req.Description, CreatedBy: actorID}
	if err := h.commission.CreateRule(item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create commission rule"})
	}
	h.insertAudit(actorID, "COMMISSION_RULE_CREATE", "commission_rule", item.RuleID, item, c.IP())
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *AdminConfigHandler) DeleteRule(c *fiber.Ctx) error {
	ruleID, err := helper.ParsePositiveInt64Param(c, "rule_id")
	if err != nil {
		return helper.BadRequest(c, "invalid rule_id")
	}
	item, err := h.commission.DeactivateRule(ruleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to deactivate commission rule"})
	}
	actorID, _ := helper.UserIDFromHeader(c)
	h.insertAudit(actorID, "COMMISSION_RULE_DEACTIVATE", "commission_rule", ruleID, item, c.IP())
	return c.JSON(fiber.Map{"rule_id": ruleID, "effective_to": item.EffectiveTo})
}

func (h *AdminConfigHandler) ListExemptions(c *fiber.Ctx) error {
	items, err := h.commission.ListExemptions(!strings.EqualFold(c.Query("active_only", "true"), "false"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch commission exemptions"})
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *AdminConfigHandler) CreateExemption(c *fiber.Ctx) error {
	var req dto.CreateCommissionExemptionRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	if req.UserID <= 0 || strings.TrimSpace(req.Reason) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id and reason are required"})
	}
	if exists, _ := h.commission.ActiveExemptionExists(req.UserID); exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "user already has an active exemption"})
	}
	var expiresAt *time.Time
	if req.ExemptTo != nil {
		if parsedTime, err := time.Parse(time.RFC3339, *req.ExemptTo); err != nil {
			return helper.BadRequest(c, "exempt_to must be RFC3339")
		} else {
			expiresAt = &parsedTime
		}
	}
	actorID, _ := helper.UserIDFromHeader(c)
	item := &domain.CommissionExemption{FactoryID: req.UserID, Reason: strings.TrimSpace(req.Reason), ExpiresAt: expiresAt, CreatedBy: actorID}
	if err := h.commission.CreateExemption(item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create commission exemption"})
	}
	h.insertAudit(actorID, "COMMISSION_EXEMPTION_CREATE", "commission_exemption", item.ExemptionID, item, c.IP())
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *AdminConfigHandler) DeleteExemption(c *fiber.Ctx) error {
	exemptionID, err := helper.ParsePositiveInt64Param(c, "exemption_id")
	if err != nil {
		return helper.BadRequest(c, "invalid exemption_id")
	}
	actorID, _ := helper.UserIDFromHeader(c)
	item, err := h.commission.RevokeExemption(exemptionID, actorID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to revoke commission exemption"})
	}
	h.insertAudit(actorID, "COMMISSION_EXEMPTION_REVOKE", "commission_exemption", exemptionID, item, c.IP())
	return c.JSON(fiber.Map{"exemption_id": exemptionID, "revoked_at": item.RevokedAt, "revoked_by": item.RevokedBy})
}

func (h *AdminConfigHandler) ListAuditLog(c *fiber.Ctx) error {
	filter := domain.AdminAuditFilter{
		Action:     strings.TrimSpace(c.Query("action")),
		TargetType: strings.TrimSpace(c.Query("target_type")),
		Page:       c.QueryInt("page", 1),
		PageSize:   c.QueryInt("page_size", 20),
	}
	actorID, err := helper.ParseOptionalPositiveInt64Query(c, "actor_id")
	if err != nil {
		return helper.BadRequest(c, "invalid actor_id")
	}
	filter.ActorID = actorID
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
	items, total, err := h.audit.List(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch audit log"})
	}
	return c.JSON(fiber.Map{"data": items, "pagination": domain.Pagination{Page: helper.MaxInt(filter.Page, 1), PageSize: helper.NormalizePageSize(filter.PageSize), Total: total}})
}

func (h *AdminConfigHandler) insertAudit(actorID int64, action, targetType string, targetID int64, payload interface{}, ip string) {
	raw, _ := json.Marshal(payload)
	_ = h.audit.Insert(&domain.AdminAuditLog{ActorID: actorID, Action: action, TargetType: targetType, TargetID: strconv.FormatInt(targetID, 10), Payload: raw, IPAddress: &ip})
}
