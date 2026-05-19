package rfq

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/helper"
	authservice "github.com/yourusername/wemake/internal/service/auth"
	quotationservice "github.com/yourusername/wemake/internal/service/quotation"
	rfqservice "github.com/yourusername/wemake/internal/service/rfq"
)

// FactoryRFQBoardHandler provides unified endpoints for the factory RFQ board
// and detail pages, consolidating multiple round-trips into single calls.
type FactoryRFQBoardHandler struct {
	rfqService       *rfqservice.RFQService
	quotationService *quotationservice.QuotationService
	auth             *authservice.AuthService
	db               *sqlx.DB
}

func NewFactoryRFQBoardHandler(
	rfqService *rfqservice.RFQService,
	quotationService *quotationservice.QuotationService,
	auth *authservice.AuthService,
	db *sqlx.DB,
) *FactoryRFQBoardHandler {
	return &FactoryRFQBoardHandler{
		rfqService:       rfqService,
		quotationService: quotationService,
		auth:             auth,
		db:               db,
	}
}

type factoryRFQBoardResponse struct {
	RFQs               interface{} `json:"rfqs"`
	FactoryCategoryIDs []int64     `json:"factory_category_ids"`
}

// GetBoard handles GET /factory/rfq-board
// Returns matching RFQs + factory's own category IDs in one call.
func (h *FactoryRFQBoardHandler) GetBoard(c *fiber.Ctx) error {
	userID, _, err := helper.RequireFactoryUser(c, h.auth)
	if err != nil {
		return err
	}

	query := helper.QueryParams(c)
	kind := query.String("kind")
	showDismissed := strings.EqualFold(query.String("show_dismissed"), "true")

	rfqs, err := h.rfqService.ListMatchingForFactory(userID, "", kind, showDismissed)
	if err != nil {
		return helper.JSONError(c, fiber.StatusInternalServerError, "failed to fetch matching rfqs")
	}

	var catIDs []int64
	if err := h.db.Select(&catIDs, `
		SELECT category_id FROM map_factory_categories WHERE factory_id = $1 ORDER BY category_id
	`, userID); err != nil {
		catIDs = []int64{}
	}
	if catIDs == nil {
		catIDs = []int64{}
	}

	return c.JSON(factoryRFQBoardResponse{
		RFQs:               rfqs,
		FactoryCategoryIDs: catIDs,
	})
}

type factoryRFQDetailResponse struct {
	RFQ        *domain.RFQ        `json:"rfq"`
	Quotations []domain.Quotation `json:"quotations"`
}

// GetDetail handles GET /factory/rfqs/:rfq_id/detail
// Returns RFQ detail + all quotations for that RFQ in one call.
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

	return c.JSON(factoryRFQDetailResponse{
		RFQ:        rfq,
		Quotations: quotations,
	})
}
