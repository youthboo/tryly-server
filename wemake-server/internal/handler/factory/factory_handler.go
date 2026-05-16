package factory

import (
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/repository"
	authservice "github.com/yourusername/wemake/internal/service/auth"
	factoryservice "github.com/yourusername/wemake/internal/service/factory"
)

const errMsgInvalidFactoryID = "invalid factory_id"

type FactoryHandler struct {
	service *factoryservice.FactoryService
	auth    *authservice.AuthService
}

func NewFactoryHandler(service *factoryservice.FactoryService, authService *authservice.AuthService) *FactoryHandler {
	return &FactoryHandler{service: service, auth: authService}
}

func (h *FactoryHandler) GetMe(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_USER_CONTEXT", "invalid user context"))
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return helper.WriteAPIError(c, helper.UnauthorizedAPIError("USER_NOT_FOUND", "user not found"))
	}
	if u.Role != domain.RoleFactory {
		return helper.WriteAPIError(c, helper.ForbiddenAPIError("FACTORY_ROLE_REQUIRED", "factory role required"))
	}
	item, err := h.service.GetPublicDetail(userID)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return helper.WriteAPIError(c, helper.NotFoundAPIError("FACTORY_NOT_FOUND", "factory profile not found"))
		}
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_FACTORY_FAILED", "failed to fetch factory"))
	}
	return helper.WriteListResponse(c, []interface{}{item}, 1)
}

func (h *FactoryHandler) List(c *fiber.Ctx) error {
	items, err := h.service.ListPublic()
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_FACTORIES_FAILED", "failed to fetch factories"))
	}
	return helper.WriteListResponse(c, items, len(items))
}

func (h *FactoryHandler) GetByID(c *fiber.Ctx) error {
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_FACTORY_ID", errMsgInvalidFactoryID))
	}
	item, err := h.service.GetPublicDetail(int64(factoryID))
	if err != nil {
		if repository.IsNotFoundError(err) {
			return helper.WriteAPIError(c, helper.NotFoundAPIError("FACTORY_NOT_FOUND", "factory not found"))
		}
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_FACTORY_FAILED", "failed to fetch factory"))
	}
	return c.JSON(item)
}

func (h *FactoryHandler) PatchProfile(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.WriteAPIError(c, helper.UnauthorizedAPIError("UNAUTHORIZED", "unauthorized"))
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_FACTORY_ID", errMsgInvalidFactoryID))
	}
	if int64(factoryID) != userID {
		return helper.WriteAPIError(c, helper.ForbiddenAPIError("FORBIDDEN", "forbidden"))
	}

	var req dto.PatchFactoryProfileRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}

	fields := map[string]interface{}{}
	if req.FactoryName != nil {
		name := helper.NormalizeString(*req.FactoryName)
		if name == "" {
			return helper.WriteAPIError(c, helper.BadRequestAPIError("EMPTY_FACTORY_NAME", "factory_name cannot be empty"))
		}
		fields["factory_name"] = name
	}
	if req.TaxID != nil {
		fields["tax_id"] = helper.NormalizeString(*req.TaxID)
	}
	if req.Description != nil {
		fields["description"] = helper.NormalizeString(*req.Description)
	}
	if req.ImageURL != nil {
		imageURL := helper.NormalizeString(*req.ImageURL)
		if imageURL == "" {
			fields["image_url"] = nil
		} else {
			fields["image_url"] = imageURL
		}
	}
	if req.BackgroundImageURL != nil {
		backgroundImageURL := helper.NormalizeString(*req.BackgroundImageURL)
		if backgroundImageURL == "" {
			fields["background_image_url"] = nil
		} else {
			fields["background_image_url"] = backgroundImageURL
		}
	}
	if req.FactoryTypeID != nil {
		if *req.FactoryTypeID <= 0 {
			return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_FACTORY_TYPE_ID", "factory_type_id must be positive"))
		}
		fields["factory_type_id"] = *req.FactoryTypeID
	}

	if err := h.service.PatchProfile(int64(factoryID), fields); err != nil {
		if repository.IsNotFoundError(err) {
			return helper.WriteAPIError(c, helper.NotFoundAPIError("FACTORY_NOT_FOUND", "factory profile not found"))
		}
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to update factory"), patchProfileErrorMap())
	}

	item, err := h.service.GetPublicDetail(int64(factoryID))
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_LATEST_FAILED", "factory updated but failed to fetch latest data"))
	}
	return c.JSON(item)
}

func (h *FactoryHandler) ListCategories(c *fiber.Ctx) error {
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_FACTORY_ID", errMsgInvalidFactoryID))
	}
	ok, err := h.service.FactoryExistsActive(int64(factoryID))
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("VERIFY_FACTORY_FAILED", "failed to verify factory"))
	}
	if !ok {
		return helper.WriteAPIError(c, helper.NotFoundAPIError("FACTORY_NOT_FOUND", "factory not found"))
	}
	items, err := h.service.ListFactoryCategories(int64(factoryID))
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_CATEGORIES_FAILED", "failed to fetch categories"))
	}
	return helper.WriteListResponse(c, items, len(items))
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
		return helper.WriteAPIError(c, helper.UnauthorizedAPIError("UNAUTHORIZED", "unauthorized"))
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_FACTORY_ID", errMsgInvalidFactoryID))
	}
	if int64(factoryID) != userID {
		return helper.WriteAPIError(c, helper.ForbiddenAPIError("FORBIDDEN", "forbidden"))
	}
	var body addFactoryCategoryBody
	if err := helper.ParseAndValidateBody(c, &body, map[string]string{
		"CategoryID": "body must include category_id (positive integer)",
	}); err != nil {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_PAYLOAD", "body must include category_id (positive integer)"))
	}
	err = h.service.AddFactoryCategory(int64(factoryID), body.CategoryID)
	if err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to add category"), addCategoryErrorMap())
	}
	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"factory_id":  factoryID,
		"category_id": body.CategoryID,
	})
}

func (h *FactoryHandler) RemoveCategory(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.WriteAPIError(c, helper.UnauthorizedAPIError("UNAUTHORIZED", "unauthorized"))
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_FACTORY_ID", errMsgInvalidFactoryID))
	}
	if int64(factoryID) != userID {
		return helper.WriteAPIError(c, helper.ForbiddenAPIError("FORBIDDEN", "forbidden"))
	}
	categoryID, err := helper.ParsePositiveInt64Param(c, "category_id")
	if err != nil || categoryID <= 0 {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_CATEGORY_ID", "invalid category_id"))
	}
	err = h.service.RemoveFactoryCategory(int64(factoryID), categoryID)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return helper.WriteAPIError(c, helper.NotFoundAPIError("MAPPING_NOT_FOUND", "mapping not found"))
		}
		return helper.WriteAPIError(c, helper.InternalServerError("REMOVE_CATEGORY_FAILED", "failed to remove category"))
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *FactoryHandler) ReplaceCategories(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.WriteAPIError(c, helper.UnauthorizedAPIError("UNAUTHORIZED", "unauthorized"))
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_FACTORY_ID", errMsgInvalidFactoryID))
	}
	if int64(factoryID) != userID {
		return helper.WriteAPIError(c, helper.ForbiddenAPIError("FORBIDDEN", "forbidden"))
	}
	var body replaceFactoryCategoriesBody
	if err := helper.ParseAndValidateBodyWithMessage(c, &body, map[string]string{
		"CategoryIDs": "body must include category_ids with at least one positive integer",
	}, "invalid payload"); err != nil {
		return err
	}
	categoryIDs, ok := validatePositiveUniqueIDs(body.CategoryIDs)
	if !ok {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_CATEGORY_IDS", "body must include category_ids with at least one positive integer"))
	}
	if err := h.service.ReplaceFactoryCategories(int64(factoryID), categoryIDs); err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to replace categories"), replaceCategoriesErrorMap())
	}
	items, err := h.service.ListFactoryCategories(int64(factoryID))
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_LATEST_FAILED", "categories updated but failed to fetch latest data"))
	}
	return c.JSON(fiber.Map{
		"factory_id": factoryID,
		"categories": items,
	})
}

func (h *FactoryHandler) ListSubCategories(c *fiber.Ctx) error {
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_FACTORY_ID", errMsgInvalidFactoryID))
	}
	ok, err := h.service.FactoryExistsActive(int64(factoryID))
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("VERIFY_FACTORY_FAILED", "failed to verify factory"))
	}
	if !ok {
		return helper.WriteAPIError(c, helper.NotFoundAPIError("FACTORY_NOT_FOUND", "factory not found"))
	}
	items, err := h.service.ListFactorySubCategories(int64(factoryID))
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_SUB_CATEGORIES_FAILED", "failed to fetch sub-categories"))
	}
	return helper.WriteListResponse(c, items, len(items))
}

type addFactorySubCategoryBody struct {
	SubCategoryID int64 `json:"sub_category_id" validate:"gt=0"`
}

func (h *FactoryHandler) AddSubCategory(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.WriteAPIError(c, helper.UnauthorizedAPIError("UNAUTHORIZED", "unauthorized"))
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_FACTORY_ID", errMsgInvalidFactoryID))
	}
	if int64(factoryID) != userID {
		return helper.WriteAPIError(c, helper.ForbiddenAPIError("FORBIDDEN", "forbidden"))
	}
	var body addFactorySubCategoryBody
	if err := helper.ParseAndValidateBody(c, &body, map[string]string{
		"SubCategoryID": "body must include sub_category_id (positive integer)",
	}); err != nil {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_PAYLOAD", "body must include sub_category_id (positive integer)"))
	}
	err = h.service.AddFactorySubCategory(int64(factoryID), body.SubCategoryID)
	if err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to add sub-category"), addSubCategoryErrorMap())
	}
	c.Status(fiber.StatusCreated)
	return c.JSON(fiber.Map{
		"factory_id":      factoryID,
		"sub_category_id": body.SubCategoryID,
	})
}

func (h *FactoryHandler) RemoveSubCategory(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.WriteAPIError(c, helper.UnauthorizedAPIError("UNAUTHORIZED", "unauthorized"))
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_FACTORY_ID", errMsgInvalidFactoryID))
	}
	if int64(factoryID) != userID {
		return helper.WriteAPIError(c, helper.ForbiddenAPIError("FORBIDDEN", "forbidden"))
	}
	subID, err := helper.ParsePositiveInt64Param(c, "sub_category_id")
	if err != nil || subID <= 0 {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_SUB_CATEGORY_ID", "invalid sub_category_id"))
	}
	err = h.service.RemoveFactorySubCategory(int64(factoryID), subID)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return helper.WriteAPIError(c, helper.NotFoundAPIError("MAPPING_NOT_FOUND", "mapping not found"))
		}
		return helper.WriteAPIError(c, helper.InternalServerError("REMOVE_SUB_CATEGORY_FAILED", "failed to remove sub-category"))
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *FactoryHandler) ReplaceSubCategories(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.WriteAPIError(c, helper.UnauthorizedAPIError("UNAUTHORIZED", "unauthorized"))
	}
	factoryID, err := c.ParamsInt("factory_id")
	if err != nil || factoryID <= 0 {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_FACTORY_ID", errMsgInvalidFactoryID))
	}
	if int64(factoryID) != userID {
		return helper.WriteAPIError(c, helper.ForbiddenAPIError("FORBIDDEN", "forbidden"))
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
			return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_SUB_CATEGORY_IDS", "sub_category_ids must contain only positive integers"))
		}
	}

	if err := h.service.ReplaceFactorySubCategories(int64(factoryID), subCategoryIDs); err != nil {
		return helper.MapServiceError(c, err, helper.ErrorMessage(fiber.StatusInternalServerError, "failed to replace sub-categories"), replaceSubCategoriesErrorMap())
	}
	items, err := h.service.ListFactorySubCategories(int64(factoryID))
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_LATEST_FAILED", "sub-categories updated but failed to fetch latest data"))
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
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_USER_CONTEXT", "invalid user context"))
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return helper.WriteAPIError(c, helper.UnauthorizedAPIError("USER_NOT_FOUND", "user not found"))
	}
	if helper.NormalizeRole(u.Role) != domain.RoleFactory {
		return helper.WriteAPIError(c, helper.ForbiddenAPIError("FACTORY_ROLE_REQUIRED", "factory role required"))
	}
	item, err := h.service.GetAnalytics(userID)
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_ANALYTICS_FAILED", "failed to fetch analytics"))
	}
	return c.JSON(item)
}

func (h *FactoryHandler) GetDashboard(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_USER_CONTEXT", "invalid user context"))
	}
	u, err := h.auth.GetUserByID(userID)
	if err != nil {
		return helper.WriteAPIError(c, helper.UnauthorizedAPIError("USER_NOT_FOUND", "user not found"))
	}
	if helper.NormalizeRole(u.Role) != domain.RoleFactory {
		return helper.WriteAPIError(c, helper.ForbiddenAPIError("FACTORY_ROLE_REQUIRED", "factory role required"))
	}
	item, err := h.service.GetDashboard(userID)
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_DASHBOARD_FAILED", "failed to fetch dashboard"))
	}
	return c.JSON(item)
}
