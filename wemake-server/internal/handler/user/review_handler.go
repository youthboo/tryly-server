package user

import (
	"database/sql"
	"github.com/yourusername/wemake/internal/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	userservice "github.com/yourusername/wemake/internal/service/user"
)

type ReviewHandler struct {
	service *userservice.ReviewService
}

func NewReviewHandler(service *userservice.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: service}
}

func (h *ReviewHandler) ListByFactory(c *fiber.Ctx) error {
	factoryID, err := helper.RequireInt64Param(c, "factory_id")
	if err != nil {
		return err
	}
	items, err := h.service.ListByFactoryID(int64(factoryID))
	if err != nil {
		return helper.JSONInternal(c, "failed to sort reviews")
	}
	return c.JSON(items)
}

func (h *ReviewHandler) GetSummaryByFactory(c *fiber.Ctx) error {
	factoryID, err := helper.RequireInt64Param(c, "factory_id")
	if err != nil {
		return err
	}
	item, err := h.service.GetSummaryByFactoryID(int64(factoryID))
	if err != nil {
		return helper.JSONInternal(c, "failed to summarize reviews")
	}
	return c.JSON(item)
}

func (h *ReviewHandler) Create(c *fiber.Ctx) error {
	factoryID, err := helper.RequireInt64Param(c, "factory_id")
	if err != nil {
		return err
	}
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}

	var req domain.FactoryReview
	if err := helper.ParseBody(c, &req, "invalid payload"); err != nil {
		return err
	}
	req.FactoryID = int64(factoryID)
	req.UserID = userID

	if err := h.service.Create(&req); err != nil {
		if err == userservice.ErrReviewImagesInvalid {
			return helper.BadRequestError(c, err.Error())
		}
		return helper.JSONInternal(c, "failed to create review")
	}
	return c.Status(fiber.StatusCreated).JSON(req)
}

func (h *ReviewHandler) UpdateByUser(c *fiber.Ctx) error {
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}
	reviewID, err := helper.RequireInt64Param(c, "review_id")
	if err != nil {
		return err
	}
	var req struct {
		Rating    int                `json:"rating"`
		Comment   string             `json:"comment"`
		ImageURLs domain.StringArray `json:"image_urls"`
	}
	if err := helper.ParseBody(c, &req, "invalid payload"); err != nil {
		return err
	}
	item, err := h.service.UpdateByUser(int64(reviewID), userID, req.Rating, req.Comment, req.ImageURLs)
	if err != nil {
		if err == userservice.ErrReviewImagesInvalid {
			return helper.BadRequestError(c, err.Error())
		}
		if err == sql.ErrNoRows {
			return helper.BadRequestError(c, "review cannot be edited")
		}
		return helper.JSONInternal(c, "failed to update review")
	}
	return c.JSON(item)
}

func (h *ReviewHandler) DeleteByUser(c *fiber.Ctx) error {
	userID, err := helper.RequireAuthenticatedUserID(c)
	if err != nil {
		return err
	}
	reviewID, err := helper.RequireInt64Param(c, "review_id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteByUser(int64(reviewID), userID); err != nil {
		if err == sql.ErrNoRows {
			return helper.BadRequestError(c, "review cannot be deleted")
		}
		return helper.JSONInternal(c, "failed to delete review")
	}
	return c.JSON(fiber.Map{"message": "review deleted"})
}
