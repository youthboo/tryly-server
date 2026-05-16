package wallet

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
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
	type reqBody struct {
		Amount float64 `json:"amount" validate:"gt=0"`
	}
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req reqBody
	if err := helper.ParseAndValidateBody(c, &req, map[string]string{
		"Amount": "amount must be greater than 0",
	}); err != nil {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid intent_id"})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid intent_id"})
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
