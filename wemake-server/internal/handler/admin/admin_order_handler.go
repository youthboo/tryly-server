package admin

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/repository"
	adminrepo "github.com/yourusername/wemake/internal/repository/admin"
	orderservice "github.com/yourusername/wemake/internal/service/order"
)

type AdminOrderHandler struct {
	repo            *adminrepo.AdminOrderRepository
	service         *orderservice.OrderService
	withdrawal      *repository.WithdrawalRepository
	adminWithdrawal *adminrepo.AdminWithdrawalRepository
	dispute         *repository.DisputeRepository
	adminDispute    *adminrepo.AdminDisputeRepository
	audit           *adminrepo.AdminAuditRepository
}

func NewAdminOrderHandler(repo *adminrepo.AdminOrderRepository, service *orderservice.OrderService, withdrawal *repository.WithdrawalRepository, adminWithdrawal *adminrepo.AdminWithdrawalRepository, dispute *repository.DisputeRepository, adminDispute *adminrepo.AdminDisputeRepository, audit *adminrepo.AdminAuditRepository) *AdminOrderHandler {
	return &AdminOrderHandler{repo: repo, service: service, withdrawal: withdrawal, adminWithdrawal: adminWithdrawal, dispute: dispute, adminDispute: adminDispute, audit: audit}
}

func (h *AdminOrderHandler) List(c *fiber.Ctx) error {
	filter := domain.AdminOrderFilter{
		Status:   strings.TrimSpace(c.Query("status")),
		Search:   strings.TrimSpace(c.Query("search")),
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("page_size", 20),
	}
	factoryID, err := parseOptionalPositiveInt64Query(c, "factory_id")
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid factory_id")
	}
	filter.FactoryID = factoryID
	userID, err := parseOptionalPositiveInt64Query(c, "user_id")
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid user_id")
	}
	filter.UserID = userID
	dateFrom, err := parseOptionalDateQuery(c, "date_from")
	if err != nil {
		return badRequest(c, "date_from must be YYYY-MM-DD")
	}
	filter.DateFrom = dateFrom
	dateTo, err := parseOptionalDateQuery(c, "date_to")
	if err != nil {
		return badRequest(c, "date_to must be YYYY-MM-DD")
	}
	filter.DateTo = dateTo
	items, total, err := h.repo.ListAdmin(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch orders"})
	}
	return c.JSON(fiber.Map{"data": items, "pagination": domain.Pagination{Page: maxInt(filter.Page, 1), PageSize: normalizePageSize(filter.PageSize), Total: total}})
}

func (h *AdminOrderHandler) GetByID(c *fiber.Ctx) error {
	orderID, err := parsePositiveInt64Param(c, "order_id")
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid order_id")
	}
	detail, err := h.service.GetAdminDetailByID(orderID)
	if err != nil {
		return writeServiceError(c, err, "failed to fetch order", notFoundCase(errNotFound, "order not found"))
	}
	finance, err := h.repo.GetAdminFinance(orderID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch order finance"})
	}
	return c.JSON(domain.AdminOrderDetailResponse{OrderDetailResponse: detail, AdminFinance: *finance})
}

func (h *AdminOrderHandler) PatchStatus(c *fiber.Ctx) error {
	orderID, err := parsePositiveInt64Param(c, "order_id")
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid order_id")
	}
	var req struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := requireBody(c, &req); err != nil {
		return err
	}
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	valid := map[string]struct{}{"PP": {}, "PR": {}, "WF": {}, "QC": {}, "SH": {}, "DL": {}, "AC": {}, "CP": {}, "CC": {}}
	if _, ok := valid[status]; !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid status"})
	}
	if err := h.service.UpdateStatus(orderID, status, nil); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update order status"})
	}
	actorID, _ := getUserIDFromHeader(c)
	payload, _ := json.Marshal(map[string]interface{}{"status": status, "reason": req.Reason})
	ip := c.IP()
	_ = h.audit.Insert(&domain.AdminAuditLog{ActorID: actorID, Action: "ORDER_STATUS_CHANGE", TargetType: "order", TargetID: strconv.FormatInt(orderID, 10), Payload: payload, IPAddress: &ip})
	return c.JSON(fiber.Map{"order_id": orderID, "status": status})
}

func (h *AdminOrderHandler) ListWithdrawals(c *fiber.Ctx) error {
	factoryID, err := parseOptionalPositiveInt64Query(c, "factory_id")
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid factory_id")
	}
	items, total, err := h.adminWithdrawal.ListAdmin(strings.TrimSpace(c.Query("status")), factoryID, c.QueryInt("page", 1), c.QueryInt("page_size", 20))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch withdrawals"})
	}
	return c.JSON(fiber.Map{"data": items, "pagination": domain.Pagination{Page: maxInt(c.QueryInt("page", 1), 1), PageSize: normalizePageSize(c.QueryInt("page_size", 20)), Total: total}})
}

func (h *AdminOrderHandler) PatchWithdrawal(c *fiber.Ctx) error {
	requestID, err := parsePositiveInt64Param(c, "request_id")
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid request_id")
	}
	var req struct {
		Status string  `json:"status"`
		Note   *string `json:"note"`
	}
	if err := requireBody(c, &req); err != nil {
		return err
	}
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status != "AP" && status != "RJ" && status != "CP" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be AP, RJ, or CP"})
	}
	if err := h.withdrawal.UpdateStatus(requestID, status, req.Note); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update withdrawal"})
	}
	actorID, _ := getUserIDFromHeader(c)
	payload, _ := json.Marshal(map[string]interface{}{"status": status, "note": req.Note})
	ip := c.IP()
	_ = h.audit.Insert(&domain.AdminAuditLog{ActorID: actorID, Action: "WITHDRAWAL_STATUS_CHANGE", TargetType: "withdrawal", TargetID: strconv.FormatInt(requestID, 10), Payload: payload, IPAddress: &ip})
	return c.JSON(fiber.Map{"request_id": requestID, "status": status, "processed_at": time.Now().UTC()})
}

func (h *AdminOrderHandler) ListDisputes(c *fiber.Ctx) error {
	orderID, err := parseOptionalPositiveInt64Query(c, "order_id")
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid order_id")
	}
	items, total, err := h.adminDispute.ListAdmin(strings.TrimSpace(c.Query("status")), orderID, c.QueryInt("page", 1), c.QueryInt("page_size", 20))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch disputes"})
	}
	return c.JSON(fiber.Map{"data": items, "pagination": domain.Pagination{Page: maxInt(c.QueryInt("page", 1), 1), PageSize: normalizePageSize(c.QueryInt("page_size", 20)), Total: total}})
}

func (h *AdminOrderHandler) PatchDispute(c *fiber.Ctx) error {
	disputeID, err := parsePositiveInt64Param(c, "dispute_id")
	if err != nil {
		return jsonError(c, fiber.StatusBadRequest, "invalid dispute_id")
	}
	var req struct {
		Status     string  `json:"status"`
		Resolution *string `json:"resolution"`
	}
	if err := requireBody(c, &req); err != nil {
		return err
	}
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status != "RS" && status != "CL" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be RS or CL"})
	}
	if err := h.dispute.UpdateStatus(disputeID, status, req.Resolution); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update dispute"})
	}
	actorID, _ := getUserIDFromHeader(c)
	payload, _ := json.Marshal(map[string]interface{}{"status": status, "resolution": req.Resolution})
	ip := c.IP()
	_ = h.audit.Insert(&domain.AdminAuditLog{ActorID: actorID, Action: "DISPUTE_STATUS_CHANGE", TargetType: "dispute", TargetID: strconv.FormatInt(disputeID, 10), Payload: payload, IPAddress: &ip})
	item, _ := h.dispute.GetByID(disputeID)
	return c.JSON(item)
}
