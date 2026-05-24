package rfq

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/helper"
	platformrepo "github.com/yourusername/wemake/internal/repository/platform_config"
	authservice "github.com/yourusername/wemake/internal/service/auth"
	quotationservice "github.com/yourusername/wemake/internal/service/quotation"
	rfqservice "github.com/yourusername/wemake/internal/service/rfq"
)

type FactoryRFQBoardHandler struct {
	rfqService       *rfqservice.RFQService
	quotationService *quotationservice.QuotationService
	auth             *authservice.AuthService
	platformConfig   *platformrepo.PlatformConfigRepository
}

func NewFactoryRFQBoardHandler(
	rfqService *rfqservice.RFQService,
	quotationService *quotationservice.QuotationService,
	auth *authservice.AuthService,
	platformConfig *platformrepo.PlatformConfigRepository,
) *FactoryRFQBoardHandler {
	return &FactoryRFQBoardHandler{
		rfqService:       rfqService,
		quotationService: quotationService,
		auth:             auth,
		platformConfig:   platformConfig,
	}
}

func (h *FactoryRFQBoardHandler) GetBoard(c *fiber.Ctx) error {
	userID, _, err := helper.RequireFactoryUser(c, h.auth)
	if err != nil {
		return err
	}

	query := helper.QueryParams(c)
	kind := query.String("kind")
	showDismissed := strings.EqualFold(query.String("show_dismissed"), "true")

	resp, err := h.rfqService.GetFactoryBoard(userID, "", kind, showDismissed)
	if err != nil {
		return helper.JSONError(c, fiber.StatusInternalServerError, "failed to fetch matching rfqs")
	}
	return c.JSON(resp)
}

type commissionConfigPayload struct {
	VatRate        float64 `json:"vat_rate"`
	CommissionRate float64 `json:"commission_rate"`
}

type factoryRFQDetailResponse struct {
	RFQ              *domain.RFQ             `json:"rfq"`
	Quotations       []domain.Quotation      `json:"quotations"`
	CommissionConfig commissionConfigPayload `json:"commission_config"`
}

func (h *FactoryRFQBoardHandler) GetDetail(c *fiber.Ctx) error {
	userID, _, err := helper.RequireFactoryUser(c, h.auth)
	if err != nil {
		return err
	}
	rfqID, err := helper.RequireInt64Param(c, "rfq_id")
	if err != nil {
		return err
	}

	rfq, err := h.rfqService.GetForViewer(userID, domain.RoleFactory, rfqID)
	if err != nil {
		return helper.JSONError(c, fiber.StatusNotFound, "rfq not found")
	}

	quotations, err := h.quotationService.ListByRFQID(rfqID)
	if err != nil {
		quotations = []domain.Quotation{}
	}
	if quotations == nil {
		quotations = []domain.Quotation{}
	}

	commCfg := commissionConfigPayload{VatRate: 7, CommissionRate: 5}
	if cfg, cfgErr := h.platformConfig.GetByFactoryID(userID); cfgErr == nil && cfg != nil {
		commCfg.VatRate = cfg.VatRate
		commCfg.CommissionRate = cfg.DefaultCommissionRate
	}

	return c.JSON(factoryRFQDetailResponse{
		RFQ:              rfq,
		Quotations:       quotations,
		CommissionConfig: commCfg,
	})
}
