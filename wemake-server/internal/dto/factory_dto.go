package dto

// Factory Request DTOs
type PatchFactoryProfileRequest struct {
	FactoryName        *string `json:"factory_name"`
	TaxID              *string `json:"tax_id"`
	Description        *string `json:"description"`
	FactoryTypeID      *int64  `json:"factory_type_id"`
	ImageURL           *string `json:"image_url"`
	BackgroundImageURL *string `json:"background_image_url"`
}

type AddCategoryRequest struct {
	CategoryID    int64 `json:"category_id" validate:"gt=0"`
	SubCategoryID *int64 `json:"sub_category_id"`
}

type ReplaceCategoriesRequest struct {
	CategoryIDs []int64 `json:"category_ids"`
}

type AddSubCategoryRequest struct {
	CategoryID    int64 `json:"category_id" validate:"gt=0"`
	SubCategoryID int64 `json:"sub_category_id" validate:"gt=0"`
}

type ReplaceSubCategoriesRequest struct {
	SubCategoryIDs []int64 `json:"sub_category_ids"`
}

type AssignFactoryConfigRequest struct {
	ConfigID int64 `json:"config_id" validate:"gt=0"`
}
