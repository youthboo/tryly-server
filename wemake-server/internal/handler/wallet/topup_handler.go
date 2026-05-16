package wallet

import (
	"errors"
	"github.com/yourusername/wemake/internal/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/repository"
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
		if repository.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wallet not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create topup intent"})
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
		if repository.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "topup intent not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch topup intent"})
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
		if errors.Is(err, walletservice.ErrTopupAlreadyProcessed) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		if repository.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "topup intent not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to confirm topup"})
	}
	return c.JSON(intent)
}
