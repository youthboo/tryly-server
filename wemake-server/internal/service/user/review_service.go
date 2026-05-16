package user

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/domainutil"
	"github.com/yourusername/wemake/internal/helper"
	userrepo "github.com/yourusername/wemake/internal/repository/user"
)

const maxReviewImages = 5

type ReviewService struct {
	repo *userrepo.ReviewRepository
}

func NewReviewService(repo *userrepo.ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

func (s *ReviewService) ListByFactoryID(factoryID int64) ([]domain.FactoryReview, error) {
	return s.repo.ListByFactoryID(factoryID)
}

func (s *ReviewService) Create(review *domain.FactoryReview) error {
	review.ImageURLs = domain.StringArray(domainutil.NormalizeStringSlice([]string(review.ImageURLs)))
	if len(review.ImageURLs) > maxReviewImages {
		return ErrReviewImagesInvalid
	}
	return s.repo.Create(review)
}

func (s *ReviewService) GetSummaryByFactoryID(factoryID int64) (*domain.FactoryReviewSummary, error) {
	return s.repo.GetSummaryByFactoryID(factoryID)
}

func (s *ReviewService) UpdateByUser(reviewID, userID int64, rating int, comment string, imageURLs domain.StringArray) (*domain.FactoryReview, error) {
	if rating < 1 || rating > 5 || strings.TrimSpace(comment) == "" || len(strings.TrimSpace(comment)) > 2000 {
		return nil, sql.ErrNoRows
	}
	imageURLs = domain.StringArray(domainutil.NormalizeStringSlice([]string(imageURLs)))
	if len(imageURLs) > maxReviewImages {
		return nil, ErrReviewImagesInvalid
	}
	return s.repo.UpdateByUser(reviewID, userID, rating, comment, imageURLs)
}

func (s *ReviewService) DeleteByUser(reviewID, userID int64) error {
	item, err := s.repo.SoftDeleteByUser(reviewID, userID)
	if err != nil {
		return err
	}
	return helper.WithTx(context.Background(), s.repo.DB(), func(tx *sqlx.Tx) error {
		return s.repo.SyncFactoryAggregateTx(tx, item.FactoryID)
	})
}
