package factory

import (
	"errors"

	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/repository"
	factoryrepo "github.com/yourusername/wemake/internal/repository/factory"
	"github.com/yourusername/wemake/internal/service"
	factoryservice "github.com/yourusername/wemake/internal/service/factory"
)

const errMsgInvalidFactoryID = "invalid factory_id"

type FactoryHandler struct {
	service *factoryservice.FactoryService
	auth    *service.AuthService
}

func NewFactoryHandler(service *factoryservice.FactoryService, authService *service.AuthService) *FactoryHandler {
	return &FactoryHandler{service: service, auth: authService}
}

func (h *FactoryHandler) GetMe(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user context"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	if u.Role != domain.RoleFactory {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "factory role required"})
	}
	item, err := h.service.GetPublicDetail(userID)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "factory profile not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch factory"})
	}
	return c.JSON(item)
}

func (h *FactoryHandler) List(c *fiber.Ctx) error {
	items, err := h.service.ListPublic()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch factories"})
	}
	return c.JSON(items)
}

func (h *FactoryHandler) GetByID(c *fiber.Ctx) error {
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errMsgInvalidFactoryID})
	}
	item, err := h.service.GetPublicDetail(int64(factoryID))
	if err != nil {
		if repository.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "factory not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch factory"})
	}
	return c.JSON(item)
}

func (h *FactoryHandler) PatchProfile(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errMsgInvalidFactoryID})
	}
	if int64(factoryID) != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}

	var req dto.PatchFactoryProfileRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}

	fields := map[string]interface{}{}
	if req.FactoryName != nil {
		name := strings.TrimSpace(*req.FactoryName)
		if name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "factory_name cannot be empty"})
		}
		fields["factory_name"] = name
	}
	if req.TaxID != nil {
		fields["tax_id"] = strings.TrimSpace(*req.TaxID)
	}
	if req.Description != nil {
		fields["description"] = strings.TrimSpace(*req.Description)
	}
	if req.ImageURL != nil {
		imageURL := strings.TrimSpace(*req.ImageURL)
		if imageURL == "" {
			fields["image_url"] = nil
		} else {
			fields["image_url"] = imageURL
		}
	}
	if req.BackgroundImageURL != nil {
		backgroundImageURL := strings.TrimSpace(*req.BackgroundImageURL)
		if backgroundImageURL == "" {
			fields["background_image_url"] = nil
		} else {
			fields["background_image_url"] = backgroundImageURL
		}
	}
	if req.FactoryTypeID != nil {
		if *req.FactoryTypeID <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "factory_type_id must be positive"})
		}
		fields["factory_type_id"] = *req.FactoryTypeID
	}

	if err := h.service.PatchProfile(int64(factoryID), fields); err != nil {
		if repository.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "factory profile not found"})
		}
		if errors.Is(err, factoryrepo.ErrInvalidFactoryType) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid factory_type_id"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update factory"})
	}

	item, err := h.service.GetPublicDetail(int64(factoryID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "factory updated but failed to fetch latest data"})
	}
	return c.JSON(item)
}

func (h *FactoryHandler) ListCategories(c *fiber.Ctx) error {
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errMsgInvalidFactoryID})
	}
	ok, err := h.service.FactoryExistsActive(int64(factoryID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to verify factory"})
	}
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "factory not found"})
	}
	items, err := h.service.ListFactoryCategories(int64(factoryID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch categories"})
	}
	return c.JSON(items)
}

type addFactoryCategoryBody struct {
	CategoryID int64 `json:"category_id" validate:"gt=0"`
}

type replaceFactoryCategoriesBody struct {
	CategoryIDs []int64 `json:"category_ids" validate:"min=1,dive,gt=0"`
}

type replaceFactorySubCategoriesBody struct {
	SubCategoryIDs []int64 `json:"sub_category_ids" validate:"omitempty,dive,gt=0"`
}

func validatePositiveUniqueIDs(ids []int64) ([]int64, bool) {
	if len(ids) == 0 {
		return nil, false
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, false
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, true
}

func (h *FactoryHandler) AddCategory(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errMsgInvalidFactoryID})
	}
	if int64(factoryID) != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
	var body addFactoryCategoryBody
	if err := helper.ParseAndValidateBody(c, &body, map[string]string{
		"CategoryID": "body must include category_id (positive integer)",
	}); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "body must include category_id (positive integer)"})
	}
	err = h.service.AddFactoryCategory(int64(factoryID), body.CategoryID)
	if err != nil {
		if errors.Is(err, factoryrepo.ErrDuplicateFactoryCategory) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "category already linked to this factory"})
		}
		if errors.Is(err, factoryrepo.ErrInvalidFactoryCategory) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid category_id"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to add category"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"factory_id":  factoryID,
		"category_id": body.CategoryID,
	})
}

func (h *FactoryHandler) RemoveCategory(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errMsgInvalidFactoryID})
	}
	if int64(factoryID) != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
	categoryID, err := helper.ParsePositiveInt64Param(c, "category_id")
	if err != nil || categoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid category_id"})
	}
	err = h.service.RemoveFactoryCategory(int64(factoryID), categoryID)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "mapping not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to remove category"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *FactoryHandler) ReplaceCategories(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errMsgInvalidFactoryID})
	}
	if int64(factoryID) != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
	var body replaceFactoryCategoriesBody
	if err := helper.ParseAndValidateBodyWithMessage(c, &body, map[string]string{
		"CategoryIDs": "body must include category_ids with at least one positive integer",
	}, "invalid payload"); err != nil {
		return err
	}
	categoryIDs, ok := validatePositiveUniqueIDs(body.CategoryIDs)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "body must include category_ids with at least one positive integer"})
	}
	if err := h.service.ReplaceFactoryCategories(int64(factoryID), categoryIDs); err != nil {
		if errors.Is(err, factoryrepo.ErrInvalidFactoryCategory) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid category_id"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to replace categories"})
	}
	items, err := h.service.ListFactoryCategories(int64(factoryID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "categories updated but failed to fetch latest data"})
	}
	return c.JSON(fiber.Map{
		"factory_id": factoryID,
		"categories": items,
	})
}

func (h *FactoryHandler) ListSubCategories(c *fiber.Ctx) error {
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errMsgInvalidFactoryID})
	}
	ok, err := h.service.FactoryExistsActive(int64(factoryID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to verify factory"})
	}
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "factory not found"})
	}
	items, err := h.service.ListFactorySubCategories(int64(factoryID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch sub-categories"})
	}
	return c.JSON(items)
}

type addFactorySubCategoryBody struct {
	SubCategoryID int64 `json:"sub_category_id" validate:"gt=0"`
}

func (h *FactoryHandler) AddSubCategory(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errMsgInvalidFactoryID})
	}
	if int64(factoryID) != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
	var body addFactorySubCategoryBody
	if err := helper.ParseAndValidateBody(c, &body, map[string]string{
		"SubCategoryID": "body must include sub_category_id (positive integer)",
	}); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "body must include sub_category_id (positive integer)"})
	}
	err = h.service.AddFactorySubCategory(int64(factoryID), body.SubCategoryID)
	if err != nil {
		if errors.Is(err, factoryrepo.ErrDuplicateFactorySubCategory) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "sub-category already linked"})
		}
		if errors.Is(err, factoryrepo.ErrInvalidFactorySubCategory) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid sub_category_id"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to add sub-category"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"factory_id":      factoryID,
		"sub_category_id": body.SubCategoryID,
	})
}

func (h *FactoryHandler) RemoveSubCategory(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errMsgInvalidFactoryID})
	}
	if int64(factoryID) != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
	subID, err := helper.ParsePositiveInt64Param(c, "sub_category_id")
	if err != nil || subID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid sub_category_id"})
	}
	err = h.service.RemoveFactorySubCategory(int64(factoryID), subID)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "mapping not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to remove sub-category"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *FactoryHandler) ReplaceSubCategories(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errMsgInvalidFactoryID})
	}
	if int64(factoryID) != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
	var body replaceFactorySubCategoriesBody
	if err := helper.ParseAndValidateBodyWithMessage(c, &body, map[string]string{
		"SubCategoryIDs": "sub_category_ids must contain only positive integers",
	}, "invalid payload"); err != nil {
		return err
	}

	subCategoryIDs := make([]int64, 0, len(body.SubCategoryIDs))
	if len(body.SubCategoryIDs) > 0 {
		var ok bool
		subCategoryIDs, ok = validatePositiveUniqueIDs(body.SubCategoryIDs)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "sub_category_ids must contain only positive integers"})
		}
	}

	if err := h.service.ReplaceFactorySubCategories(int64(factoryID), subCategoryIDs); err != nil {
		if errors.Is(err, factoryrepo.ErrInvalidFactorySubCategory) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid sub_category_id"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to replace sub-categories"})
	}
	items, err := h.service.ListFactorySubCategories(int64(factoryID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "sub-categories updated but failed to fetch latest data"})
	}
	return c.JSON(fiber.Map{
		"factory_id":     factoryID,
		"sub_categories": items,
	})
}

// GET /factories/me/analytics
func (h *FactoryHandler) GetAnalytics(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user context"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	if strings.TrimSpace(strings.ToUpper(u.Role)) != domain.RoleFactory {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "factory role required"})
	}
	item, err := h.service.GetAnalytics(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch analytics"})
	}
	return c.JSON(item)
}

func (h *FactoryHandler) GetDashboard(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user context"})
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}
	if strings.TrimSpace(strings.ToUpper(u.Role)) != domain.RoleFactory {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "factory role required"})
	}
	item, err := h.service.GetDashboard(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch dashboard"})
	}
	return c.JSON(item)
}
