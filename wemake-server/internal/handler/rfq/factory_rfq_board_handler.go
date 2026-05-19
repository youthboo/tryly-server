package rfq

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	authservice "github.com/yourusername/wemake/internal/service/auth"
	rfqservice "github.com/yourusername/wemake/internal/service/rfq"
	"github.com/yourusername/wemake/internal/helper"
)

// FactoryRFQBoardHandler serves GET /factory/rfq-board — a single endpoint that
// returns both the matching RFQ list and the factory's own category IDs so that
// the front-end only needs one round-trip to render the board page.
type FactoryRFQBoardHandler struct {
	rfqService *rfqservice.RFQService
	auth       *authservice.AuthService
	db         *sqlx.DB
}

func NewFactoryRFQBoardHandler(rfqService *rfqservice.RFQService, auth *authservice.AuthService, db *sqlx.DB) *FactoryRFQBoardHandler {
	return &FactoryRFQBoardHandler{rfqService: rfqService, auth: auth, db: db}
}

type factoryRFQBoardResponse struct {
	RFQs              interface{} `json:"rfqs"`
	FactoryCategoryIDs []int64    `json:"factory_category_ids"`
}

// GetBoard handles GET /factory/rfq-board
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
		RFQs:              rfqs,
		FactoryCategoryIDs: catIDs,
	})
}
