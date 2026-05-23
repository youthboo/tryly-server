package user

import (
	"github.com/yourusername/wemake/internal/domain"
	userrepo "github.com/yourusername/wemake/internal/repository/user"
)

type FavoriteService struct {
	repo *userrepo.FavoriteRepository
}

func NewFavoriteService(repo *userrepo.FavoriteRepository) *FavoriteService {
	return &FavoriteService{repo: repo}
}

func (s *FavoriteService) ListByUserID(userID int64) ([]domain.Favorite, error) {
	return s.repo.ListByUserID(userID)
}

func (s *FavoriteService) Add(fav *domain.Favorite) error {
	return s.repo.Add(fav)
}

func (s *FavoriteService) Remove(userID, showcaseID int64) error {
	return s.repo.Remove(userID, showcaseID)
}
