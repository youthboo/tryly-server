package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	adminrepo "github.com/yourusername/wemake/internal/repository/admin"
)

type AdminCustomerHandler struct {
	customers   *adminrepo.CustomerAdminRepository
	settlements *adminrepo.SettlementAdminRepository
}

func NewAdminCustomerHandler(
	customers *adminrepo.CustomerAdminRepository,
	settlements *adminrepo.SettlementAdminRepository,
) *AdminCustomerHandler {
	return &AdminCustomerHandler{customers: customers, settlements: settlements}
}

// GET /api/admin/customers
func (h *AdminCustomerHandler) ListCustomers(c *fiber.Ctx) error {
	query := helper.QueryParams(c)
	search := query.String("search")
	limit, offset := helper.LimitOffset(c, helper.DefaultPageSize)

	var isActive *bool
	if v := query.String("is_active"); v != "" {
		b := v == "true" || v == "1"
		isActive = &b
	}

	items, total, err := h.customers.ListCustomers(search, isActive, limit, offset)
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch customers")
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
	userID, err := helper.ParsePositiveInt64Param(c, "user_id")
	if err != nil {
		return helper.JSONError(c, fiber.StatusBadRequest, "invalid user_id")
	}

	detail, err := h.customers.GetCustomerDetail(userID)
	if err != nil {
		return helper.WriteServiceError(c, err, "failed to fetch customer", helper.NotFoundCase(helper.ErrNotFound, "customer not found"))
	}
	return c.JSON(detail)
}

// GET /api/admin/customers/:user_id/wallet
func (h *AdminCustomerHandler) GetCustomerWallet(c *fiber.Ctx) error {
	userID, err := helper.ParsePositiveInt64Param(c, "user_id")
	if err != nil {
		return helper.JSONError(c, fiber.StatusBadRequest, "invalid user_id")
	}

	wallet, err := h.customers.GetCustomerWallet(userID)
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch wallet")
	}
	return c.JSON(wallet)
}

// GET /api/admin/customers/:user_id/orders
func (h *AdminCustomerHandler) ListCustomerOrders(c *fiber.Ctx) error {
	userID, err := helper.ParsePositiveInt64Param(c, "user_id")
	if err != nil {
		return helper.JSONError(c, fiber.StatusBadRequest, "invalid user_id")
	}
	limit, offset := helper.LimitOffset(c, helper.DefaultPageSize)

	items, total, err := h.customers.ListCustomerOrders(userID, limit, offset)
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch orders")
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
	query := helper.QueryParams(c)
	limit := helper.ClampInt(query.Int("limit", 5), helper.MinPageSize, helper.MaxPageSize)
	items, err := h.customers.ListTopCustomers(limit)
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch top customers")
	}
	return c.JSON(fiber.Map{"top_customers": items})
}

// GET /api/admin/factories/:factory_id/settlements
func (h *AdminCustomerHandler) ListFactorySettlements(c *fiber.Ctx) error {
	factoryID, err := helper.ParsePositiveInt64Param(c, "factory_id")
	if err != nil {
		return helper.JSONError(c, fiber.StatusBadRequest, "invalid factory_id")
	}
	limit, offset := helper.LimitOffset(c, helper.DefaultPageSize)

	items, total, err := h.settlements.ListByFactory(factoryID, limit, offset)
	if err != nil {
		return helper.JSONInternal(c, "failed to fetch settlements")
	}
	return c.JSON(fiber.Map{
		"settlements": items,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}
