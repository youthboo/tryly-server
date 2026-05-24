package notification

import (
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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
	query := `SELECT noti_id, user_id, type, title,
			COALESCE(message, '') AS message, COALESCE(link_to, '') AS link_to,
			is_read, read_at, deleted_at, created_at
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
		"COALESCE(message, '') AS message",
		"COALESCE(link_to, '') AS link_to",
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
	err := r.db.Get(&count, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE AND deleted_at IS NULL AND type != 'CHAT_MESSAGE'`, userID)
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

// ListWithFilter returns paginated notifications filtered by type, plus global unread_count for badge.
// filterTypes nil/empty = all types. Results are ordered: unread first, then newest.
func (r *NotificationRepository) ListWithFilter(userID int64, filterTypes []string, limit, offset int) (items []domain.Notification, total int64, unreadCount int64, err error) {
	const base = `SELECT noti_id, user_id, type, title,
		COALESCE(message,'') AS message, COALESCE(link_to,'') AS link_to,
		is_read, read_at, deleted_at, created_at
		FROM notifications WHERE user_id = $1 AND deleted_at IS NULL AND type != 'CHAT_MESSAGE'`

	if len(filterTypes) > 0 {
		if err = r.db.Select(&items, base+` AND type = ANY($2::text[])
			ORDER BY is_read ASC, created_at DESC LIMIT $3 OFFSET $4`,
			userID, pq.Array(filterTypes), limit, offset); err != nil {
			log.Printf("ListWithFilter Select (filtered) error: %v", err)
			return
		}
		if err = r.db.Get(&total,
			`SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND deleted_at IS NULL AND type=ANY($2::text[])`,
			userID, pq.Array(filterTypes)); err != nil {
			log.Printf("ListWithFilter Count (filtered) error: %v", err)
			return
		}
	} else {
		if err = r.db.Select(&items, base+`
			ORDER BY is_read ASC, created_at DESC LIMIT $2 OFFSET $3`,
			userID, limit, offset); err != nil {
			log.Printf("ListWithFilter Select error: %v", err)
			return
		}
		if err = r.db.Get(&total,
			`SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND deleted_at IS NULL AND type != 'CHAT_MESSAGE'`,
			userID); err != nil {
			log.Printf("ListWithFilter Count error: %v", err)
			return
		}
	}
	unreadCount, err = r.GetUnreadCount(userID)
	return
}

// MarkAsReadReturnCount marks one notification read (idempotent) and returns the new global unread_count.
func (r *NotificationRepository) MarkAsReadReturnCount(notiID, userID int64) (int64, error) {
	_, err := r.db.Exec(
		`UPDATE notifications SET is_read=TRUE, read_at=NOW()
		 WHERE noti_id=$1 AND user_id=$2 AND is_read=FALSE`,
		notiID, userID)
	if err != nil {
		return 0, err
	}
	return r.GetUnreadCount(userID)
}

// MarkAllReadWithFilter marks unread notifications read, optionally scoped to a type list.
// Returns (updated_count, new_global_unread_count, error).
func (r *NotificationRepository) MarkAllReadWithFilter(userID int64, filterTypes []string) (updatedCount int64, unreadCount int64, err error) {
	var res interface{ RowsAffected() (int64, error) }
	if len(filterTypes) > 0 {
		res, err = r.db.Exec(
			`UPDATE notifications SET is_read=TRUE, read_at=NOW()
			 WHERE user_id=$1 AND is_read=FALSE AND deleted_at IS NULL AND type=ANY($2::text[])`,
			userID, pq.Array(filterTypes))
	} else {
		res, err = r.db.Exec(
			`UPDATE notifications SET is_read=TRUE, read_at=NOW()
			 WHERE user_id=$1 AND is_read=FALSE AND deleted_at IS NULL`,
			userID)
	}
	if err != nil {
		return
	}
	updatedCount, err = res.RowsAffected()
	if err != nil {
		return
	}
	unreadCount, err = r.GetUnreadCount(userID)
	return
}

// PollRow is the SSE poll result; includes global unread_count per row.
type PollRow struct {
	NotiID      int64      `db:"noti_id"    json:"noti_id"`
	Type        string     `db:"type"       json:"type"`
	Title       string     `db:"title"      json:"title"`
	Message     string     `db:"message"    json:"message"`
	LinkTo      string     `db:"link_to"    json:"link_to"`
	IsRead      bool       `db:"is_read"    json:"is_read"`
	CreatedAt   string     `db:"created_at" json:"created_at"`
	UnreadCount int64      `db:"unread_count" json:"unread_count"`
}

// PollNew returns notifications with noti_id > lastNotiID for SSE streaming.
func (r *NotificationRepository) PollNew(userID, lastNotiID int64) ([]PollRow, error) {
	const query = `
		SELECT n.noti_id, n.type, n.title,
			COALESCE(n.message,'') AS message,
			COALESCE(n.link_to,'') AS link_to,
			n.is_read,
			to_char(n.created_at AT TIME ZONE 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at,
			(SELECT COUNT(*) FROM notifications
			 WHERE user_id=$1 AND is_read=FALSE AND deleted_at IS NULL) AS unread_count
		FROM notifications n
		WHERE n.user_id=$1 AND n.noti_id>$2 AND n.deleted_at IS NULL
		ORDER BY n.noti_id ASC`
	var rows []PollRow
	err := r.db.Select(&rows, query, userID, lastNotiID)
	return rows, err
}
