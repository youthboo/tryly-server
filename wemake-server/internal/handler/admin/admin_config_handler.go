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
	items, err := h.commission.ListRules(factoryID, !strings.EqualFold(helper.QueryString(c, "active_only"), "false"))
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch commission rules")
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *AdminConfigHandler) CreateRule(c *fiber.Ctx) error {
	var req dto.CreateCommissionRuleRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	v := domain.NewValidationCollector()
	v.AddIf(req.CommissionRate < 0 || req.CommissionRate > 100, "commission_rate", "must be between 0 and 100")
	if err := helper.ValidateRequest(c, v); err != nil {
		return err
	}
	from := time.Now().UTC()
	actorID := helper.OptionalActorID(c)
	item := &domain.CommissionRule{RatePercent: helper.MoneyDecimal(req.CommissionRate), EffectiveFrom: from, Note: req.Description, CreatedBy: actorID}
	if err := h.commission.CreateRule(item); err != nil {
		return helper.JSONInternal(c, "failed to create commission rule")
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
		return helper.JSONInternal(c, "failed to deactivate commission rule")
	}
	actorID := helper.OptionalActorID(c)
	h.insertAudit(actorID, "COMMISSION_RULE_DEACTIVATE", "commission_rule", ruleID, item, c.IP())
	return c.JSON(fiber.Map{"rule_id": ruleID, "effective_to": item.EffectiveTo})
}

func (h *AdminConfigHandler) ListExemptions(c *fiber.Ctx) error {
	items, err := h.commission.ListExemptions(!strings.EqualFold(helper.QueryString(c, "active_only"), "false"))
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch commission exemptions")
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *AdminConfigHandler) CreateExemption(c *fiber.Ctx) error {
	var req dto.CreateCommissionExemptionRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	v := domain.NewValidationCollector()
	v.AddIf(req.UserID <= 0, "user_id", "is required")
	v.AddIf(strings.TrimSpace(req.Reason) == "", "reason", "is required")
	if err := helper.ValidateRequest(c, v); err != nil {
		return err
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
	actorID := helper.OptionalActorID(c)
	item := &domain.CommissionExemption{FactoryID: req.UserID, Reason: strings.TrimSpace(req.Reason), ExpiresAt: expiresAt, CreatedBy: actorID}
	if err := h.commission.CreateExemption(item); err != nil {
		return helper.JSONInternal(c, "failed to create commission exemption")
	}
	h.insertAudit(actorID, "COMMISSION_EXEMPTION_CREATE", "commission_exemption", item.ExemptionID, item, c.IP())
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *AdminConfigHandler) DeleteExemption(c *fiber.Ctx) error {
	exemptionID, err := helper.ParsePositiveInt64Param(c, "exemption_id")
	if err != nil {
		return helper.BadRequest(c, "invalid exemption_id")
	}
	actorID := helper.OptionalActorID(c)
	item, err := h.commission.RevokeExemption(exemptionID, actorID)
	if err != nil {
		return helper.JSONInternal(c, "failed to revoke commission exemption")
	}
	h.insertAudit(actorID, "COMMISSION_EXEMPTION_REVOKE", "commission_exemption", exemptionID, item, c.IP())
	return c.JSON(fiber.Map{"exemption_id": exemptionID, "revoked_at": item.RevokedAt, "revoked_by": item.RevokedBy})
}

func (h *AdminConfigHandler) ListAuditLog(c *fiber.Ctx) error {
	query := helper.QueryParams(c)
	page, pageSize := query.PageSize(helper.DefaultPageSize)
	filter := domain.AdminAuditFilter{
		Action:     query.String("action"),
		TargetType: query.String("target_type"),
		Page:       page,
		PageSize:   pageSize,
		ActorID:    query.OptionalPositiveInt64("actor_id"),
		DateFrom:   query.OptionalDate("date_from"),
		DateTo:     query.OptionalDate("date_to"),
	}
	if err := query.Err(); err != nil {
		return err
	}
	items, total, err := h.audit.List(filter)
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch audit log")
	}
	return helper.PaginatedResponse(c, items, filter.Page, filter.PageSize, total)
}

func (h *AdminConfigHandler) insertAudit(actorID int64, action, targetType string, targetID int64, payload interface{}, ip string) {
	raw, _ := json.Marshal(payload)
	_ = h.audit.Insert(&domain.AdminAuditLog{ActorID: actorID, Action: action, TargetType: targetType, TargetID: strconv.FormatInt(targetID, 10), Payload: raw, IPAddress: &ip})
}
