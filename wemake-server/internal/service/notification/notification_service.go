package notification

import (
	"github.com/yourusername/wemake/internal/domain"
	notificationrepo "github.com/yourusername/wemake/internal/repository/notification"
)

// re-export PollRow so callers don't need to import the repo package
type PollRow = notificationrepo.PollRow

type NotificationService struct {
	repo *notificationrepo.NotificationRepository
}

func NewNotificationService(repo *notificationrepo.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) ListByUserID(userID int64) ([]domain.Notification, error) {
	return s.repo.ListByUserID(userID)
}

func (s *NotificationService) MarkAsRead(notiID, userID int64) error {
	return s.repo.MarkAsRead(notiID, userID)
}

func (s *NotificationService) Create(noti *domain.Notification) error {
	return s.repo.Create(noti)
}

func (s *NotificationService) ListPaginated(userID int64, page, limit int, unreadOnly bool) ([]domain.Notification, int64, int64, error) {
	return s.repo.ListPaginated(userID, page, limit, unreadOnly)
}

func (s *NotificationService) GetUnreadCount(userID int64) (int64, error) {
	return s.repo.GetUnreadCount(userID)
}

func (s *NotificationService) MarkAllRead(userID int64) (int64, error) {
	return s.repo.MarkAllRead(userID)
}

func (s *NotificationService) SoftDelete(notiID, userID int64) error {
	return s.repo.SoftDelete(notiID, userID)
}

func (s *NotificationService) ListWithFilter(userID int64, filterTypes []string, limit, offset int) ([]domain.Notification, int64, int64, error) {
	return s.repo.ListWithFilter(userID, filterTypes, limit, offset)
}

func (s *NotificationService) MarkAsReadReturnCount(notiID, userID int64) (int64, error) {
	return s.repo.MarkAsReadReturnCount(notiID, userID)
}

func (s *NotificationService) MarkAllReadWithFilter(userID int64, filterTypes []string) (int64, int64, error) {
	return s.repo.MarkAllReadWithFilter(userID, filterTypes)
}

func (s *NotificationService) PollNew(userID, lastNotiID int64) ([]notificationrepo.PollRow, error) {
	return s.repo.PollNew(userID, lastNotiID)
}
