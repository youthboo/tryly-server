package notification

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/domain"
)

type NotificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) ListByUserID(userID int64) ([]domain.Notification, error) {
	var items []domain.Notification
	query := `SELECT noti_id, user_id, type, title, message, link_to, is_read, read_at,
			NULL::jsonb AS data, NULL::bigint AS reference_id, deleted_at, created_at
		FROM notifications WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	err := r.db.Select(&items, query, userID)
	return items, err
}

func (r *NotificationRepository) MarkAsRead(notiID, userID int64) error {
	query := `UPDATE notifications SET is_read = TRUE, read_at = NOW() WHERE noti_id = $1 AND user_id = $2 AND deleted_at IS NULL`
	_, err := r.db.Exec(query, notiID, userID)
	return err
}

func (r *NotificationRepository) Create(noti *domain.Notification) error {
	query := `
		INSERT INTO notifications (user_id, type, title, message, link_to)
		VALUES (:user_id, :type, :title, :message, :link_to)
		RETURNING noti_id, created_at, is_read, read_at
	`
	rows, err := r.db.NamedQuery(query, noti)
	if err != nil {
		return err
	}
	if rows.Next() {
		err = rows.Scan(&noti.NotiID, &noti.CreatedAt, &noti.IsRead, &noti.ReadAt)
	}
	rows.Close()
	return err
}

func (r *NotificationRepository) ListPaginated(userID int64, page, limit int, unreadOnly bool) ([]domain.Notification, int64, int64, error) {
	offset := (page - 1) * limit

	conditions := sq.And{
		sq.Eq{"user_id": userID},
		sq.Eq{"deleted_at": nil},
	}
	if unreadOnly {
		conditions = append(conditions, sq.Eq{"is_read": false})
	}

	countQuery := sq.Select("COUNT(*)").
		From("notifications").
		Where(conditions)

	var total int64
	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, 0, err
	}
	if err := r.db.Get(&total, countSQL, countArgs...); err != nil {
		return nil, 0, 0, err
	}

	var unreadCount int64
	unreadQuery := sq.Select("COUNT(*)").
		From("notifications").
		Where(sq.And{
			sq.Eq{"user_id": userID},
			sq.Eq{"is_read": false},
			sq.Eq{"deleted_at": nil},
		})
	unreadSQL, unreadArgs, err := unreadQuery.ToSql()
	if err != nil {
		return nil, 0, 0, err
	}
	if err := r.db.Get(&unreadCount, unreadSQL, unreadArgs...); err != nil {
		return nil, 0, 0, err
	}

	query := sq.Select(
		"noti_id",
		"user_id",
		"type",
		"title",
		"message",
		"link_to",
		"is_read",
		"read_at",
		"NULL::jsonb AS data",
		"NULL::bigint AS reference_id",
		"deleted_at",
		"created_at",
	).
		From("notifications").
		Where(conditions).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, 0, 0, err
	}

	var items []domain.Notification
	if err := r.db.Select(&items, sql, args...); err != nil {
		return nil, 0, 0, err
	}

	return items, total, unreadCount, nil
}

func (r *NotificationRepository) GetUnreadCount(userID int64) (int64, error) {
	var count int64
	err := r.db.Get(&count, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE AND deleted_at IS NULL`, userID)
	return count, err
}

func (r *NotificationRepository) MarkAllRead(userID int64) (int64, error) {
	res, err := r.db.Exec(`UPDATE notifications SET is_read = TRUE, read_at = NOW() WHERE user_id = $1 AND is_read = FALSE AND deleted_at IS NULL`, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *NotificationRepository) SoftDelete(notiID, userID int64) error {
	_, err := r.db.Exec(`UPDATE notifications SET deleted_at = NOW() WHERE noti_id = $1 AND user_id = $2 AND deleted_at IS NULL`, notiID, userID)
	return err
}
