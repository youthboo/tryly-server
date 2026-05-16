package wallet

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	walletservice "github.com/yourusername/wemake/internal/service/wallet"
)

type TopupHandler struct {
	service *walletservice.TopupService
}

func NewTopupHandler(svc *walletservice.TopupService) *TopupHandler {
	return &TopupHandler{service: svc}
}

// POST /wallets/topup
func (h *TopupHandler) CreateIntent(c *fiber.Ctx) error {
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}
	var req dto.TopupIntentRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}
	intent, err := h.service.CreateIntent(userID, req.Amount)
	if err != nil {
		return helper.MapServiceError(c, err, topupCreateFallback, topupCreateResponses)
	}
	return c.Status(fiber.StatusCreated).JSON(intent)
}

// GET /wallets/topup/:intent_id
func (h *TopupHandler) GetIntent(c *fiber.Ctx) error {
	intentID := c.Params("intent_id")
	if intentID == "" {
		return helper.BadRequestError(c, "invalid intent_id")
	}
	intent, err := h.service.GetIntent(intentID)
	if err != nil {
		return helper.MapServiceError(c, err, topupGetFallback, topupGetResponses)
	}
	return c.JSON(intent)
}

// POST /wallets/topup/:intent_id/confirm
func (h *TopupHandler) ConfirmIntent(c *fiber.Ctx) error {
	intentID := c.Params("intent_id")
	if intentID == "" {
		return helper.BadRequestError(c, "invalid intent_id")
	}
	intent, err := h.service.ConfirmIntent(intentID)
	if err != nil {
		return helper.MapServiceError(c, err, topupConfirmFallback, topupConfirmResponses)
	}
	return c.JSON(intent)
}

var topupCreateFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to create topup intent")

var topupCreateResponses = map[error]helper.ErrorResponse{
	sql.ErrNoRows: helper.ErrorMessage(fiber.StatusNotFound, "wallet not found"),
}

var topupGetFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to fetch topup intent")

var topupGetResponses = map[error]helper.ErrorResponse{
	sql.ErrNoRows: helper.ErrorMessage(fiber.StatusNotFound, "topup intent not found"),
}

var topupConfirmFallback = helper.ErrorMessage(fiber.StatusInternalServerError, "failed to confirm topup")

var topupConfirmResponses = map[error]helper.ErrorResponse{
	walletservice.ErrTopupAlreadyProcessed: helper.ErrorMessage(fiber.StatusBadRequest, walletservice.ErrTopupAlreadyProcessed.Error()),
	sql.ErrNoRows:                          helper.ErrorMessage(fiber.StatusNotFound, "topup intent not found"),
}
