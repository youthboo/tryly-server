package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/repository"
)

type CreateOrderReviewInput struct {
	Rating    int
	Comment   string
	ImageURLs domain.StringArray
}

func (s *OrderService) GetReviewState(orderID, userID int64, role string) (*domain.OrderReviewState, error) {
	if role != domain.RoleCustomer {
		return nil, domain.ErrForbidden
	}
	order, err := s.repo.GetByParticipant(orderID, userID, role)
	if err != nil {
		return nil, err
	}

	factoryName := fmt.Sprintf("โรงงาน #%d", order.FactoryID)
	if detail, detailErr := s.repo.GetDetailByParticipant(orderID, userID, role); detailErr == nil && strings.TrimSpace(detail.FactoryName) != "" {
		factoryName = detail.FactoryName
	}

	state := &domain.OrderReviewState{
		OrderID:         order.OrderID,
		FactoryID:       order.FactoryID,
		FactoryName:     factoryName,
		Eligible:        false,
		AlreadyReviewed: false,
	}

	review, err := s.reviews.GetByOrderAndUser(orderID, userID)
	if err == nil {
		state.AlreadyReviewed = true
		state.Review = review
		reason := "already_reviewed"
		state.Reason = &reason
		return state, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if normalizeOrderStatus(order.Status) != "CP" {
		reason := "order_not_completed"
		state.Reason = &reason
		return state, nil
	}

	state.Eligible = true
	return state, nil
}

func (s *OrderService) CreateReview(orderID, userID int64, role string, input CreateOrderReviewInput) (*domain.FactoryReview, error) {
	if role != domain.RoleCustomer {
		return nil, domain.ErrForbidden
	}
	if input.Rating < 1 || input.Rating > 5 {
		return nil, ErrReviewRatingInvalid
	}
	comment := strings.TrimSpace(input.Comment)
	if comment == "" || len(comment) > 1000 {
		return nil, ErrReviewCommentInvalid
	}
	imageURLs := normalizeReviewImageURLs(input.ImageURLs)
	if len(imageURLs) > maxReviewImages {
		return nil, ErrReviewImagesInvalid
	}

	order, err := s.repo.GetByParticipant(orderID, userID, role)
	if err != nil {
		return nil, err
	}
	if normalizeOrderStatus(order.Status) != "CP" {
		return nil, ErrReviewOrderNotCompleted
	}
	if _, err := s.reviews.GetByOrderAndUser(orderID, userID); err == nil {
		return nil, ErrReviewAlreadyExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	orderIDPtr := order.OrderID
	review := &domain.FactoryReview{
		FactoryID: order.FactoryID,
		UserID:    userID,
		OrderID:   &orderIDPtr,
		Rating:    input.Rating,
		Comment:   comment,
		ImageURLs: imageURLs,
	}
	if err := WithTx(context.Background(), s.db, func(tx *sqlx.Tx) error {
		if err := s.reviews.CreateForOrderTx(tx, review); err != nil {
			if errors.Is(err, repository.ErrReviewAlreadyExists) {
				return ErrReviewAlreadyExists
			}
			return err
		}
		if err := s.reviews.SyncFactoryAggregateTx(tx, order.FactoryID); err != nil {
			return err
		}
		return s.repo.InsertActivityTx(tx, orderID, &userID, "REVIEW_CREATED", map[string]interface{}{
			"review_id": review.ReviewID,
			"rating":    review.Rating,
		})
	}); err != nil {
		return nil, err
	}
	createNotificationSafe(s.notifications, &domain.Notification{
		UserID:  order.FactoryID,
		Type:    "REVIEW_RECEIVED",
		Title:   "ได้รับรีวิวใหม่",
		Message: fmt.Sprintf("ลูกค้าให้ %d ดาว: \"%s\"", review.Rating, trimNotificationPreview(review.Comment, 80)),
		LinkTo:  orderLink(orderID),
		Data: notificationData(map[string]interface{}{
			"review_id": review.ReviewID,
			"order_id":  orderID,
			"rating":    review.Rating,
			"url":       orderLink(orderID),
		}),
		ReferenceID: &review.ReviewID,
		CreatedAt:   time.Now(),
	})
	return review, nil
}
