package frontend

import (
	"database/sql"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/yourusername/wemake/internal/domain"
)

type SessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

type sessionUserRow struct {
	ID             int64           `db:"id"`
	Role           string          `db:"role"`
	FirstName      sql.NullString  `db:"first_name"`
	LastName       sql.NullString  `db:"last_name"`
	FactoryName    sql.NullString  `db:"factory_name"`
	Email          string          `db:"email"`
	Phone          sql.NullString  `db:"phone"`
	MemberSince    string          `db:"member_since"`
	Balance        float64         `db:"balance"`
	PendingBalance float64         `db:"pending_balance"`
}

type sessionRFQRow struct {
	RFQID        int64           `db:"rfq_id"`
	Title        string          `db:"title"`
	Status       string          `db:"status"`
	Quantity     sql.NullInt64   `db:"quantity"`
	TargetPrice  sql.NullFloat64 `db:"target_price"`
	CategoryName string          `db:"category_name"`
	CreatedAt    string          `db:"created_at"`
	OfferCount   int64           `db:"offer_count"`
}

type sessionOfferRow struct {
	QuoteID         int64           `db:"quote_id"`
	RFQID           int64           `db:"rfq_id"`
	FactoryID       int64           `db:"factory_id"`
	FactoryName     string          `db:"factory_name"`
	FactoryImageURL sql.NullString  `db:"factory_image_url"`
	FactoryVerified bool            `db:"factory_verified"`
	PricePerPiece   sql.NullFloat64 `db:"price_per_piece"`
	GrandTotal      sql.NullFloat64 `db:"grand_total"`
	LeadTimeDays    sql.NullInt64   `db:"lead_time_days"`
	Status          string          `db:"status"`
}

type sessionOrderRow struct {
	OrderID           int64           `db:"order_id"`
	Title             string          `db:"title"`
	FactoryID         int64           `db:"factory_id"`
	FactoryName       string          `db:"factory_name"`
	FactoryImageURL   sql.NullString  `db:"factory_image_url"`
	Status            string          `db:"status"`
	TotalAmount       sql.NullFloat64 `db:"total_amount"`
	DepositAmount     sql.NullFloat64 `db:"deposit_amount"`
	EstimatedDelivery sql.NullString  `db:"estimated_delivery"`
	CreatedAt         string          `db:"created_at"`
}

type sessionThreadRow struct {
	ConvID              int64          `db:"conv_id"`
	CounterpartID       int64          `db:"counterpart_id"`
	CounterpartName     string         `db:"counterpart_name"`
	CounterpartImageURL sql.NullString `db:"counterpart_image_url"`
	LastMessage         sql.NullString `db:"last_message"`
	LastMessageAt       string         `db:"last_message_at"`
	Unread              int64          `db:"unread"`
}

func (r *SessionRepository) GetSession(userID int64) (*domain.SessionResponse, error) {
	var (
		wg sync.WaitGroup

		userRow    *sessionUserRow
		rfqRows    []sessionRFQRow
		offerRows  []sessionOfferRow
		orderRows  []sessionOrderRow
		threadRows []sessionThreadRow
		favIDs     []int64

		userErr    error
		rfqErr     error
		offerErr   error
		orderErr   error
		threadErr  error
		favErr     error
	)

	wg.Add(6)

	go func() {
		defer wg.Done()
		var row sessionUserRow
		err := r.db.Get(&row, `
			SELECT
				u.user_id AS id,
				u.role,
				c.first_name,
				c.last_name,
				fp.factory_name,
				u.email,
				u.phone,
				TO_CHAR(u.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS member_since,
				COALESCE(w.good_fund, 0)::float8    AS balance,
				COALESCE(w.pending_fund, 0)::float8 AS pending_balance
			FROM users u
			LEFT JOIN customers c ON c.user_id = u.user_id
			LEFT JOIN factory_profiles fp ON fp.user_id = u.user_id
			LEFT JOIN wallets w ON w.user_id = u.user_id
			WHERE u.user_id = $1 AND u.is_active = TRUE
		`, userID)
		if err == nil {
			userRow = &row
		}
		userErr = err
	}()

	go func() {
		defer wg.Done()
		rfqErr = r.db.Select(&rfqRows, `
			SELECT
				r.rfq_id,
				r.title,
				r.status,
				r.quantity,
				r.target_price,
				c.name AS category_name,
				TO_CHAR(r.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at,
				COUNT(q.quote_id) AS offer_count
			FROM rfqs r
			JOIN lbi_categories c ON c.category_id = r.category_id
			LEFT JOIN quotations q ON q.rfq_id = r.rfq_id
			WHERE r.user_id = $1 AND r.status != 'CC'
			GROUP BY r.rfq_id, r.title, r.status, r.quantity, r.target_price, c.name, r.created_at
			ORDER BY r.created_at DESC
			LIMIT 10
		`, userID)
	}()

	go func() {
		defer wg.Done()
		offerErr = r.db.Select(&offerRows, `
			SELECT
				q.quote_id,
				q.rfq_id,
				q.factory_id,
				fp.factory_name,
				fp.image_url                     AS factory_image_url,
				(fp.approval_status = 'AP')      AS factory_verified,
				q.price_per_piece,
				((q.price_per_piece * r.quantity) + COALESCE(q.mold_cost, 0)) AS grand_total,
				q.lead_time_days,
				q.status
			FROM quotations q
			JOIN rfqs r ON r.rfq_id = q.rfq_id
			JOIN factory_profiles fp ON fp.user_id = q.factory_id
			WHERE r.user_id = $1 AND r.status != 'CC'
			ORDER BY q.rfq_id, q.create_time DESC
		`, userID)
	}()

	go func() {
		defer wg.Done()
		orderErr = r.db.Select(&orderRows, `
			SELECT
				o.order_id,
				rfq.title,
				o.factory_id,
				fp.factory_name,
				fp.image_url AS factory_image_url,
				o.status,
				o.total_amount,
				o.deposit_amount,
				TO_CHAR(
					COALESCE(
						o.estimated_delivery::timestamp,
						(o.created_at + (q.lead_time_days || ' days')::interval)
					), 'YYYY-MM-DD'
				) AS estimated_delivery,
				TO_CHAR(o.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at
			FROM orders o
			JOIN quotations q ON q.quote_id = o.quote_id
			JOIN rfqs rfq ON rfq.rfq_id = q.rfq_id
			JOIN factory_profiles fp ON fp.user_id = o.factory_id
			WHERE o.user_id = $1 AND o.status NOT IN ('CP', 'CN', 'CL')
			ORDER BY o.created_at DESC
			LIMIT 10
		`, userID)
	}()

	go func() {
		defer wg.Done()
		threadErr = r.db.Select(&threadRows, `
			SELECT
				c.conv_id,
				c.factory_id                  AS counterpart_id,
				fp.factory_name               AS counterpart_name,
				fp.image_url                  AS counterpart_image_url,
				c.last_message,
				TO_CHAR(c.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS last_message_at,
				COALESCE(c.unread_customer, 0) AS unread
			FROM conversations c
			JOIN factory_profiles fp ON fp.user_id = c.factory_id
			WHERE c.customer_id = $1
			ORDER BY c.updated_at DESC
			LIMIT 20
		`, userID)
	}()

	go func() {
		defer wg.Done()
		favErr = r.db.Select(&favIDs, `
			SELECT showcase_id FROM favorites WHERE user_id = $1 ORDER BY created_at DESC
		`, userID)
	}()

	wg.Wait()

	if userErr != nil {
		return nil, userErr
	}

	// Build currentUser + wallet
	name := strings.TrimSpace(strings.Join([]string{
		userRow.FirstName.String, userRow.LastName.String,
	}, " "))
	if userRow.FactoryName.Valid && userRow.FactoryName.String != "" {
		name = userRow.FactoryName.String
	}
	if name == "" {
		name = userRow.Email
	}

	currentUser := &domain.SessionCurrentUser{
		ID:          userRow.ID,
		Role:        userRow.Role,
		Name:        name,
		Email:       userRow.Email,
		Phone:       userRow.Phone.String,
		MemberSince: userRow.MemberSince,
	}
	wallet := &domain.SessionWallet{
		Balance:        userRow.Balance,
		PendingBalance: userRow.PendingBalance,
	}

	// Group offers by rfq_id
	offersByRFQ := make(map[int64][]domain.SessionOffer)
	if offerErr == nil {
		for _, o := range offerRows {
			offer := domain.SessionOffer{
				QuoteID:         o.QuoteID,
				FactoryID:       o.FactoryID,
				FactoryName:     o.FactoryName,
				FactoryVerified: o.FactoryVerified,
				Status:          o.Status,
			}
			if o.FactoryImageURL.Valid {
				s := o.FactoryImageURL.String
				offer.FactoryImageURL = &s
			}
			if o.PricePerPiece.Valid {
				v := o.PricePerPiece.Float64
				offer.PricePerPiece = &v
			}
			if o.GrandTotal.Valid {
				v := o.GrandTotal.Float64
				offer.GrandTotal = &v
			}
			if o.LeadTimeDays.Valid {
				v := o.LeadTimeDays.Int64
				offer.LeadTimeDays = &v
			}
			offersByRFQ[o.RFQID] = append(offersByRFQ[o.RFQID], offer)
		}
	}

	// Map RFQs
	rfqs := make([]domain.SessionRFQ, 0, len(rfqRows))
	if rfqErr == nil {
		for _, r := range rfqRows {
			rfq := domain.SessionRFQ{
				RFQID:        r.RFQID,
				Title:        r.Title,
				Status:       r.Status,
				CategoryName: r.CategoryName,
				CreatedAt:    r.CreatedAt,
				OfferCount:   r.OfferCount,
				Offers:       offersByRFQ[r.RFQID],
			}
			if rfq.Offers == nil {
				rfq.Offers = []domain.SessionOffer{}
			}
			if r.Quantity.Valid {
				v := r.Quantity.Int64
				rfq.Quantity = &v
			}
			if r.TargetPrice.Valid {
				v := r.TargetPrice.Float64
				rfq.TargetPrice = &v
			}
			rfqs = append(rfqs, rfq)
		}
	}

	// Map orders
	orders := make([]domain.SessionOrder, 0, len(orderRows))
	if orderErr == nil {
		for _, o := range orderRows {
			order := domain.SessionOrder{
				OrderID:     o.OrderID,
				Title:       o.Title,
				FactoryID:   o.FactoryID,
				FactoryName: o.FactoryName,
				Status:      o.Status,
				CreatedAt:   o.CreatedAt,
			}
			if o.FactoryImageURL.Valid {
				s := o.FactoryImageURL.String
				order.FactoryImageURL = &s
			}
			if o.TotalAmount.Valid {
				v := o.TotalAmount.Float64
				order.TotalAmount = &v
			}
			if o.DepositAmount.Valid {
				v := o.DepositAmount.Float64
				order.DepositAmount = &v
			}
			if o.EstimatedDelivery.Valid {
				v := o.EstimatedDelivery.String
				order.EstimatedDelivery = &v
			}
			orders = append(orders, order)
		}
	}

	// Map threads
	threads := make([]domain.SessionThread, 0, len(threadRows))
	if threadErr == nil {
		for _, t := range threadRows {
			thread := domain.SessionThread{
				ConvID:        t.ConvID,
				CounterpartID: t.CounterpartID,
				CounterpartName: t.CounterpartName,
				LastMessage:   t.LastMessage.String,
				LastMessageAt: t.LastMessageAt,
				Unread:        t.Unread,
			}
			if t.CounterpartImageURL.Valid {
				s := t.CounterpartImageURL.String
				thread.CounterpartImageURL = &s
			}
			threads = append(threads, thread)
		}
	}

	// Favorites
	favorites := favIDs
	if favorites == nil || favErr != nil {
		favorites = []int64{}
	}

	// Ensure pq.Array decoded correctly (favIDs populated via Select which handles []int64 natively)
	_ = pq.Array // keep import

	return &domain.SessionResponse{
		CurrentUser: currentUser,
		Wallet:      wallet,
		RFQs:        rfqs,
		Orders:      orders,
		Threads:     threads,
		Favorites:   favorites,
	}, nil
}
