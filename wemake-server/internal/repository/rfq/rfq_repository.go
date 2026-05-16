package rfq

import (
	"database/sql"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/domainutil"
	"github.com/yourusername/wemake/internal/helper"
)

type RFQRepository struct {
	db *sqlx.DB
}

func NewRFQRepository(db *sqlx.DB) *RFQRepository {
	return &RFQRepository{db: db}
}

func nullableDecimalToFloat64(d *decimal.Decimal) interface{} {
	if d == nil {
		return nil
	}
	f := helper.DecimalToFloat(*d)
	return &f
}

func (r *RFQRepository) DB() *sqlx.DB {
	return r.db
}

func (r *RFQRepository) Create(rfq *domain.RFQ) error {
	return r.createWithExecutor(r.db, rfq)
}

func (r *RFQRepository) CreateTx(tx *sqlx.Tx, rfq *domain.RFQ) error {
	return r.createWithExecutor(tx, rfq)
}

type rfqQueryRowExecutor interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}

const rfqSelectColumns = `
		rfq_id, user_id, COALESCE(category_id, 0) AS category_id, sub_category_id, title, quantity, details,
		0::bigint AS address_id, shipping_method_id, status, COALESCE(request_kind, 'PR') AS request_kind,
		NULL::timestamp AS uploaded_at, created_at, updated_at,
		material_grade, target_price, target_lead_time_days, NULL::date AS required_delivery_date, delivery_address_id,
		certifications_required, FALSE AS sample_required, NULL::integer AS sample_qty, NULL::text AS inspection_type,
		NULL::bigint AS conversation_id, reference_images, 'RFQ'::text AS rfq_type, 'buyer'::text AS initiated_by,
		NULL::bigint AS factory_user_id, NULL::bigint AS source_showcase_id, NULL::bigint AS source_conv_id,
		NULL::text AS boq_currency, NULL::numeric AS boq_subtotal, NULL::numeric AS boq_discount_amount,
		NULL::numeric AS boq_vat_percent, NULL::numeric AS boq_vat_amount, NULL::numeric AS boq_grand_total,
		NULL::integer AS boq_moq, NULL::integer AS boq_lead_time_days, NULL::text AS boq_payment_terms,
		NULL::integer AS boq_validity_days, NULL::text AS boq_note, NULL::timestamptz AS boq_sent_at,
		NULL::timestamptz AS boq_responded_at, NULL::text AS boq_response, NULL::text AS boq_decline_reason
`

const rfqSelectColumnsR = `
		r.rfq_id, r.user_id, COALESCE(r.category_id, 0) AS category_id, r.sub_category_id, r.title, r.quantity, r.details,
		0::bigint AS address_id, r.shipping_method_id, r.status, COALESCE(r.request_kind, 'PR') AS request_kind,
		NULL::timestamp AS uploaded_at, r.created_at, r.updated_at,
		r.material_grade, r.target_price, r.target_lead_time_days, NULL::date AS required_delivery_date, r.delivery_address_id,
		r.certifications_required, FALSE AS sample_required, NULL::integer AS sample_qty, NULL::text AS inspection_type,
		NULL::bigint AS conversation_id, r.reference_images, 'RFQ'::text AS rfq_type, 'buyer'::text AS initiated_by,
		NULL::bigint AS factory_user_id, NULL::bigint AS source_showcase_id, NULL::bigint AS source_conv_id,
		NULL::text AS boq_currency, NULL::numeric AS boq_subtotal, NULL::numeric AS boq_discount_amount,
		NULL::numeric AS boq_vat_percent, NULL::numeric AS boq_vat_amount, NULL::numeric AS boq_grand_total,
		NULL::integer AS boq_moq, NULL::integer AS boq_lead_time_days, NULL::text AS boq_payment_terms,
		NULL::integer AS boq_validity_days, NULL::text AS boq_note, NULL::timestamptz AS boq_sent_at,
		NULL::timestamptz AS boq_responded_at, NULL::text AS boq_response, NULL::text AS boq_decline_reason
`

func (r *RFQRepository) createWithExecutor(exec rfqQueryRowExecutor, rfq *domain.RFQ) error {
	query := `
		INSERT INTO rfqs (
			user_id, category_id, sub_category_id, title, quantity, details,
			shipping_method_id, status, request_kind, created_at, updated_at,
			material_grade, target_price, target_lead_time_days, delivery_address_id,
			certifications_required, reference_images
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17
		)
		RETURNING rfq_id
	`
	return exec.QueryRow(
		query,
		rfq.UserID,
		domainutil.NullablePositiveInt64(rfq.CategoryID),
		domainutil.NullableInt64(rfq.SubCategoryID),
		rfq.Title,
		rfq.Quantity,
		rfq.Details,
		domainutil.NullableInt64(rfq.ShippingMethodID),
		rfq.Status,
		domainutil.NormalizeUpperOrDefault(rfq.RequestKind, domain.RequestKindProduction),
		rfq.CreatedAt,
		rfq.UpdatedAt,
		domainutil.NullableString(rfq.MaterialGrade),
		nullableDecimalToFloat64(rfq.TargetPrice),
		domainutil.NullableInt(rfq.TargetLeadTimeDays),
		domainutil.NullableInt64(rfq.DeliveryAddressID),
		rfq.CertificationsRequired,
		rfq.ReferenceImages,
	).Scan(&rfq.RFQID)
}

func (r *RFQRepository) ListByUserID(userID int64, status string) ([]domain.RFQ, error) {
	var rfqs []domain.RFQ
	query := `
		SELECT ` + rfqSelectColumns + `
		FROM rfqs
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"
	err := r.db.Select(&rfqs, query, args...)
	if err != nil {
		return rfqs, err
	}
	for i := range rfqs {
		if err := r.enrichRFQLookups(&rfqs[i]); err != nil {
			return rfqs, err
		}
		domain.EnrichRFQBudgetFields(&rfqs[i])
	}
	return rfqs, nil
}

func (r *RFQRepository) GetByID(userID, rfqID int64) (*domain.RFQ, error) {
	var rfq domain.RFQ
	query := `
		SELECT ` + rfqSelectColumns + `
		FROM rfqs
		WHERE user_id = $1 AND rfq_id = $2
	`
	if err := r.db.Get(&rfq, query, userID, rfqID); err != nil {
		return nil, err
	}
	if rfq.AddressID > 0 {
		addr, err := r.getAddressByID(rfq.AddressID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if err == nil {
			rfq.Address = addr
		}
	}
	if err := r.enrichRFQLookups(&rfq); err != nil {
		return nil, err
	}
	domain.EnrichRFQBudgetFields(&rfq)
	return &rfq, nil
}

func (r *RFQRepository) Cancel(userID, rfqID int64) error {
	return helper.WithTx(nil, r.db, func(tx *sqlx.Tx) error {
		if _, err := tx.Exec("UPDATE rfqs SET status = 'CC', updated_at = NOW() WHERE user_id = $1 AND rfq_id = $2", userID, rfqID); err != nil {
			return err
		}
		if _, err := tx.Exec("UPDATE quotations SET status = 'EX', log_timestamp = NOW() WHERE rfq_id = $1 AND status = 'PD'", rfqID); err != nil {
			return err
		}
		return nil
	})
}

// CloseOpenRFQForUserTx sets RFQ status from OP to CL when the customer awards an order (same transaction as order create).
func (r *RFQRepository) CloseOpenRFQForUserTx(tx *sqlx.Tx, rfqID, userID int64) error {
	_, err := tx.Exec(`
		UPDATE rfqs
		SET status = 'CL', updated_at = NOW()
		WHERE rfq_id = $1 AND user_id = $2 AND status = 'OP'
	`, rfqID, userID)
	return err
}

// CloseRFQ lets a customer manually close (stop accepting new quotes) an open RFQ.
// Only OP RFQs owned by userID can be closed; pending quotations are NOT expired so
// existing accepted factories remain active.
func (r *RFQRepository) CloseRFQ(userID, rfqID int64) error {
	_, err := r.db.Exec(`
		UPDATE rfqs
		SET status = 'CL', updated_at = NOW()
		WHERE rfq_id = $1 AND user_id = $2 AND status = 'OP'
	`, rfqID, userID)
	return err
}

func (r *RFQRepository) SubCategoryBelongsToCategory(subCategoryID, categoryID int64) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM lbi_sub_categories
			WHERE sub_category_id = $1
				AND category_id = $2
				AND status = '1'
		)
	`
	err := r.db.Get(&exists, query, subCategoryID, categoryID)
	return exists, err
}

func (r *RFQRepository) ShippingMethodExists(shippingMethodID int64) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM lbi_shipping_methods
			WHERE shipping_method_id = $1
				AND status = '1'
		)
	`
	err := r.db.Get(&exists, query, shippingMethodID)
	return exists, err
}

func (r *RFQRepository) CategoryScope(categoryID int64) (string, bool, error) {
	var scope sql.NullString
	err := r.db.Get(&scope, `
		SELECT COALESCE(scope, 'PD') AS scope
		FROM lbi_categories
		WHERE category_id = $1
	`, categoryID)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if !scope.Valid || strings.TrimSpace(scope.String) == "" {
		return "PD", true, nil
	}
	return domainutil.NormalizeStatus(scope.String), true, nil
}

// GetByIDAny loads RFQ by id without customer ownership check.
func (r *RFQRepository) GetByIDAny(rfqID int64) (*domain.RFQ, error) {
	var rfq domain.RFQ
	query := `
		SELECT ` + rfqSelectColumns + `
		FROM rfqs
		WHERE rfq_id = $1
	`
	if err := r.db.Get(&rfq, query, rfqID); err != nil {
		return nil, err
	}
	if rfq.AddressID > 0 {
		addr, err := r.getAddressByID(rfq.AddressID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if err == nil {
			rfq.Address = addr
		}
	}
	if err := r.enrichRFQLookups(&rfq); err != nil {
		return nil, err
	}
	domain.EnrichRFQBudgetFields(&rfq)
	return &rfq, nil
}

func (r *RFQRepository) getAddressByID(addressID int64) (*domain.Address, error) {
	var address domain.Address
	query := `
		SELECT address_id, user_id, address_type, address_detail, sub_district_id, district_id, province_id, zip_code, is_default
		FROM addresses
		WHERE address_id = $1
	`
	if err := r.db.Get(&address, query, addressID); err != nil {
		return nil, err
	}
	return &address, nil
}

// ListMatchingForFactory returns RFQs for the factory board with quotation overlay.
//
// Rules:
//  1. OP RFQs that match category/sub-category and have NOT been quoted → "ยังไม่ได้เสนอ" tab
//  2. RFQs that this factory HAS quoted (any RFQ status) AND quotation != AC → "ติดตาม BOQ" tab
//  3. Quotations with status=AC are excluded entirely (customer accepted → became an order)
//
// The `status` param is kept for backwards compat but ignored; the WHERE logic enforces the
// rules above instead.
func (r *RFQRepository) ListMatchingForFactory(factoryID int64, status string, kind string, showDismissed bool) ([]domain.RFQ, error) {
	kinds := splitRFQKinds(kind)
	if len(kinds) == 0 {
		kinds = []string{domain.RequestKindProduction, domain.RequestKindProductSample, domain.RequestKindMaterialSample}
	}
	var rfqs []domain.RFQ
	query := `
		SELECT DISTINCT
		       ` + rfqSelectColumnsR + `,
		       (frd.factory_id IS NOT NULL) AS is_dismissed,
		       frd.dismissed_at,
		       (q.quote_id IS NULL OR q.status NOT IN ('PD','AC')) AS can_dismiss,
		       -- quotation overlay for this factory
		       q.status         AS my_quote_status,
		       q.quote_id       AS my_quote_id,
		       q.price_per_piece AS my_quoted_price
		FROM rfqs r
		LEFT JOIN quotations q
		       ON q.rfq_id = r.rfq_id AND q.factory_id = $1
		LEFT JOIN lbi_sub_categories sc ON sc.sub_category_id = r.sub_category_id
		LEFT JOIN factory_rfq_dismissals frd ON frd.rfq_id = r.rfq_id AND frd.factory_id = $1
		WHERE
		  -- Rule 3: ซ่อน quotation ที่ถูกยอมรับแล้ว (กลายเป็น order)
		  COALESCE(q.status, '') != 'AC'
		  -- Rule 1 + 2: แสดง OP ที่ยังไม่เสนอ OR RFQ ที่เสนอแล้ว (ไม่ว่า rfq status จะเป็นอะไร)
		  AND (r.status = 'OP' OR q.quote_id IS NOT NULL)
		  AND COALESCE(r.request_kind, 'PR') = ANY($2)
		  AND ($3::boolean OR frd.factory_id IS NULL)
		  AND (
			(
				COALESCE(r.request_kind, 'PR') IN ('PR', 'PS')
				AND EXISTS (
					SELECT 1
					FROM map_factory_categories mfc
					WHERE mfc.category_id = r.category_id
					  AND mfc.factory_id = $1
				)
				AND (
					r.sub_category_id IS NULL
					OR COALESCE(sc.sort_order, 0) = 99
					OR EXISTS (
						SELECT 1 FROM map_factory_sub_categories ms
						WHERE ms.factory_id = $1 AND ms.sub_category_id = r.sub_category_id
					)
				)
			)
			OR (
				COALESCE(r.request_kind, 'PR') = 'MS'
				AND EXISTS (
					SELECT 1
					FROM factory_showcases fs
					INNER JOIN lbi_categories cat ON cat.category_id = fs.category_id
					WHERE fs.factory_id = $1
					  AND fs.content_type = 'MT'
					  AND fs.status = 'AC'
					  AND COALESCE(cat.scope, 'PD') = 'MT'
					  AND fs.category_id = r.category_id
				)
			)
		  )
		ORDER BY r.created_at DESC
	`
	err := r.db.Select(&rfqs, query, factoryID, pq.Array(kinds), showDismissed)
	if err != nil {
		return rfqs, err
	}
	for i := range rfqs {
		if err := r.enrichRFQLookups(&rfqs[i]); err != nil {
			return rfqs, err
		}
		domain.EnrichRFQBudgetFields(&rfqs[i])
	}
	return rfqs, nil
}

func (r *RFQRepository) EnrichFactoryDismissalState(rfq *domain.RFQ, factoryID int64) error {
	if rfq == nil {
		return nil
	}
	type row struct {
		IsDismissed bool       `db:"is_dismissed"`
		DismissedAt *time.Time `db:"dismissed_at"`
		CanDismiss  bool       `db:"can_dismiss"`
	}
	var state row
	if err := r.db.Get(&state, `
		SELECT
			(frd.factory_id IS NOT NULL) AS is_dismissed,
			frd.dismissed_at,
			NOT EXISTS (
				SELECT 1 FROM quotations q
				WHERE q.rfq_id = $2
				  AND q.factory_id = $1
				  AND q.status IN ('PD', 'AC')
			) AS can_dismiss
		FROM rfqs r
		LEFT JOIN factory_rfq_dismissals frd ON frd.rfq_id = r.rfq_id AND frd.factory_id = $1
		WHERE r.rfq_id = $2
	`, factoryID, rfq.RFQID); err != nil {
		return err
	}
	rfq.IsDismissed = state.IsDismissed
	rfq.DismissedAt = state.DismissedAt
	rfq.CanDismiss = state.CanDismiss
	return nil
}

func (r *RFQRepository) ListMatchingFactoryIDs(rfq *domain.RFQ) ([]int64, error) {
	if rfq == nil {
		return nil, nil
	}
	return r.ListMatchingFactoryIDsForKind(rfq.RequestKind, rfq.CategoryID, rfq.SubCategoryID)
}

func (r *RFQRepository) ListMatchingFactoryIDsForKind(kind string, categoryID int64, subCategoryID *int64) ([]int64, error) {
	var ids []int64
	if domainutil.NormalizeStatus(kind) == domain.RequestKindMaterialSample {
		query := `
			SELECT DISTINCT fs.factory_id
			FROM factory_showcases fs
			INNER JOIN lbi_categories cat ON cat.category_id = fs.category_id
			LEFT JOIN factory_profiles fp ON fp.user_id = fs.factory_id
			WHERE fs.content_type = 'MT'
			  AND fs.status = 'AC'
			  AND COALESCE(cat.scope, 'PD') = 'MT'
			  AND COALESCE(fp.approval_status, 'AP') <> 'SU'
			  AND fs.category_id = $1
		`
		if err := r.db.Select(&ids, query, domainutil.NullablePositiveInt64(categoryID)); err != nil {
			return nil, err
		}
		return ids, nil
	}
	query := `
		SELECT DISTINCT mfc.factory_id
		FROM map_factory_categories mfc
		LEFT JOIN lbi_sub_categories sc ON sc.sub_category_id = $2
		LEFT JOIN factory_profiles fp ON fp.user_id = mfc.factory_id
		WHERE mfc.category_id = $1
		  AND COALESCE(fp.approval_status, 'AP') <> 'SU'
		  AND (
			$2::bigint IS NULL
			OR COALESCE(sc.sort_order, 0) = 99
			OR EXISTS (
				SELECT 1
				FROM map_factory_sub_categories ms
				WHERE ms.factory_id = mfc.factory_id
				  AND ms.sub_category_id = $2
			)
		  )
	`
	if err := r.db.Select(&ids, query, domainutil.NullablePositiveInt64(categoryID), domainutil.NullableInt64(subCategoryID)); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *RFQRepository) enrichRFQLookups(rfq *domain.RFQ) error {
	if rfq == nil {
		return nil
	}

	var categoryName sql.NullString
	if err := r.db.Get(&categoryName, `SELECT name FROM lbi_categories WHERE category_id = $1`, rfq.CategoryID); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	} else if categoryName.Valid {
		rfq.CategoryName = &categoryName.String
	}

	if rfq.SubCategoryID != nil {
		var subCategoryName sql.NullString
		if err := r.db.Get(&subCategoryName, `SELECT name FROM lbi_sub_categories WHERE sub_category_id = $1`, *rfq.SubCategoryID); err != nil {
			if err != sql.ErrNoRows {
				return err
			}
		} else if subCategoryName.Valid {
			rfq.SubCategoryName = &subCategoryName.String
		}
	}

	if rfq.ShippingMethodID != nil {
		var shippingMethodName sql.NullString
		if err := r.db.Get(&shippingMethodName, `SELECT method_name FROM lbi_shipping_methods WHERE shipping_method_id = $1`, *rfq.ShippingMethodID); err != nil {
			if err != sql.ErrNoRows {
				return err
			}
		} else if shippingMethodName.Valid {
			rfq.ShippingMethodName = &shippingMethodName.String
		}
	}

	var addressSummary sql.NullString
	if rfq.AddressID <= 0 {
		return nil
	}
	if err := r.db.Get(&addressSummary, `
		SELECT TRIM(BOTH ' ' FROM CONCAT_WS(' / ',
			NULLIF(a.address_detail, ''),
			NULLIF(sd.name_th, ''),
			NULLIF(d.name_th, ''),
			NULLIF(p.name_th, ''),
			NULLIF(a.zip_code, '')
		))
		FROM addresses a
		LEFT JOIN lbi_sub_districts sd ON sd.row_id = a.sub_district_id
		LEFT JOIN lbi_districts d ON d.row_id = a.district_id
		LEFT JOIN lbi_provinces p ON p.row_id = a.province_id
		WHERE a.address_id = $1
	`, rfq.AddressID); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	} else if addressSummary.Valid {
		rfq.AddressSummary = &addressSummary.String
	}

	return nil
}

// FactoryHasMatchingCategory returns true if factory accepts RFQ's category and sub-category rules.
func (r *RFQRepository) FactoryHasMatchingCategory(factoryID int64, rfq *domain.RFQ) (bool, error) {
	var ok bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM map_factory_categories mfc
			WHERE mfc.factory_id = $1 AND mfc.category_id = $2
		)
		AND (
			$3::bigint IS NULL
			OR EXISTS (
				SELECT 1
				FROM lbi_sub_categories sc
				WHERE sc.sub_category_id = $3
				  AND COALESCE(sc.sort_order, 0) = 99
			)
			OR EXISTS (
				SELECT 1 FROM map_factory_sub_categories ms
				WHERE ms.factory_id = $1 AND ms.sub_category_id = $3
			)
		)
	`
	err := r.db.Get(&ok, query, factoryID, rfq.CategoryID, domainutil.NullableInt64(rfq.SubCategoryID))
	return ok, err
}

func (r *RFQRepository) FactoryHasQuotationOnRFQ(factoryID, rfqID int64) (bool, error) {
	var ok bool
	err := r.db.Get(&ok, `
		SELECT EXISTS (SELECT 1 FROM quotations WHERE factory_id = $1 AND rfq_id = $2)
	`, factoryID, rfqID)
	return ok, err
}

func (r *RFQRepository) Patch(userID, rfqID int64, rfq *domain.RFQ) error {
	_, err := r.db.NamedExec(`
		UPDATE rfqs
		SET category_id = :category_id,
		    sub_category_id = :sub_category_id,
		    title = :title,
		    quantity = :quantity,
		    details = :details,
		    shipping_method_id = :shipping_method_id,
		    request_kind = :request_kind,
		    material_grade = :material_grade,
		    target_price = :target_price,
		    target_lead_time_days = :target_lead_time_days,
		    delivery_address_id = :delivery_address_id,
		    certifications_required = :certifications_required,
		    reference_images = :reference_images,
		    updated_at = NOW()
		WHERE rfq_id = :rfq_id AND user_id = :user_id AND status = 'OP'
	`, rfq)
	return err
}

func (r *RFQRepository) MarkInReviewTx(tx *sqlx.Tx, rfqID, userID int64) error {
	_, err := tx.Exec(`
		UPDATE rfqs
		SET status = 'IR', updated_at = NOW()
		WHERE rfq_id = $1 AND user_id = $2 AND status IN ('OP', 'IR')
	`, rfqID, userID)
	return err
}

func splitRFQKinds(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		item := domainutil.NormalizeStatus(part)
		switch item {
		case domain.RequestKindProduction, domain.RequestKindProductSample, domain.RequestKindMaterialSample:
		default:
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func (r *RFQRepository) LinkConversationTx(tx *sqlx.Tx, rfqID, userID, convID int64) error {
	_, err := tx.Exec(`
		UPDATE rfqs
		SET updated_at = NOW()
		WHERE rfq_id = $1 AND user_id = $2
	`, rfqID, userID)
	return err
}
