package factory

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/helper"
	factoryrepo "github.com/yourusername/wemake/internal/repository/factory"
)

func patchProfileErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		factoryrepo.ErrInvalidFactoryType: helper.ErrorMessage(fiber.StatusBadRequest, "invalid factory_type_id"),
	}
}

func addCategoryErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		factoryrepo.ErrDuplicateFactoryCategory: helper.ErrorMessage(fiber.StatusConflict, "category already linked to this factory"),
		factoryrepo.ErrInvalidFactoryCategory:   helper.ErrorMessage(fiber.StatusBadRequest, "invalid category_id"),
	}
}

func replaceCategoriesErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		factoryrepo.ErrInvalidFactoryCategory: helper.ErrorMessage(fiber.StatusBadRequest, "invalid category_id"),
	}
}

func addSubCategoryErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		factoryrepo.ErrDuplicateFactorySubCategory: helper.ErrorMessage(fiber.StatusConflict, "sub-category already linked"),
		factoryrepo.ErrInvalidFactorySubCategory:   helper.ErrorMessage(fiber.StatusBadRequest, "invalid sub_category_id"),
	}
}

func replaceSubCategoriesErrorMap() map[error]helper.ErrorResponse {
	return map[error]helper.ErrorResponse{
		factoryrepo.ErrInvalidFactorySubCategory: helper.ErrorMessage(fiber.StatusBadRequest, "invalid sub_category_id"),
	}
}
