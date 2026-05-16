package handler

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/repository"
)

type AdminCustomerHandler struct {
	customers   *repository.CustomerAdminRepository
	settlements *repository.SettlementAdminRepository
}

func NewAdminCustomerHandler(
	customers *repository.CustomerAdminRepository,
	settlements *repository.SettlementAdminRepository,
) *AdminCustomerHandler {
	return &AdminCustomerHandler{customers: customers, settlements: settlements}
}

// GET /api/admin/customers
func (h *AdminCustomerHandler) ListCustomers(c *fiber.Ctx) error {
	search := c.Query("search", "")
	limit, offset := limitOffset(c, 20)

	var isActive *bool
	if v := c.Query("is_active"); v != "" {
		b := v == "true" || v == "1"
		isActive = &b
	}

	items, total, err := h.customers.ListCustomers(search, isActive, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch customers"})
	}
	return c.JSON(fiber.Map{
		"customers": items,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GET /api/admin/customers/:user_id
func (h *AdminCustomerHandler) GetCustomerDetail(c *fiber.Ctx) error {
	userID, err := parsePositiveInt64Param(c, "user_id")
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid user_id")
	}

	detail, err := h.customers.GetCustomerDetail(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "customer not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch customer"})
	}
	return c.JSON(detail)
}

// GET /api/admin/customers/:user_id/wallet
func (h *AdminCustomerHandler) GetCustomerWallet(c *fiber.Ctx) error {
	userID, err := parsePositiveInt64Param(c, "user_id")
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid user_id")
	}

	wallet, err := h.customers.GetCustomerWallet(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch wallet"})
	}
	return c.JSON(wallet)
}

// GET /api/admin/customers/:user_id/orders
func (h *AdminCustomerHandler) ListCustomerOrders(c *fiber.Ctx) error {
	userID, err := parsePositiveInt64Param(c, "user_id")
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid user_id")
	}
	limit, offset := limitOffset(c, 20)

	items, total, err := h.customers.ListCustomerOrders(userID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch orders"})
	}
	return c.JSON(fiber.Map{
		"orders": items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GET /api/admin/dashboard/top-customers
func (h *AdminCustomerHandler) ListTopCustomers(c *fiber.Ctx) error {
	limit := clampInt(c.QueryInt("limit", 5), 1, 100)
	items, err := h.customers.ListTopCustomers(limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch top customers"})
	}
	return c.JSON(fiber.Map{"top_customers": items})
}

// GET /api/admin/factories/:factory_id/settlements
func (h *AdminCustomerHandler) ListFactorySettlements(c *fiber.Ctx) error {
	factoryID, err := parsePositiveInt64Param(c, "factory_id")
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid factory_id")
	}
	limit, offset := limitOffset(c, 20)

	items, total, err := h.settlements.ListByFactory(factoryID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch settlements"})
	}
	return c.JSON(fiber.Map{
		"settlements": items,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}
