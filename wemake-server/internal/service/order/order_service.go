package order

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	orderrepo "github.com/yourusername/wemake/internal/repository/order"
	paymentrepo "github.com/yourusername/wemake/internal/repository/payment"
	quotationrepo "github.com/yourusername/wemake/internal/repository/quotation"
	rfqrepo "github.com/yourusername/wemake/internal/repository/rfq"
	userrepo "github.com/yourusername/wemake/internal/repository/user"
	walletrepo "github.com/yourusername/wemake/internal/repository/wallet"
)

var (
	ErrQuotationRejected     = errors.New("quotation was rejected")
	ErrQuotationInvalidState = errors.New("quotation must be pending or already accepted")
	ErrNotQuotationParty     = errors.New("not authorized for this quotation")
	ErrQuotationExpired      = errors.New("QUOTATION_EXPIRED")
)

var ErrShipOrderInvalid = errors.New("tracking_no and courier are required")
var ErrOrderCannotBeCancelled = errors.New("order cannot be cancelled in its current status")
var ErrInsufficientGoodFund = errors.New("insufficient good_fund balance")
var ErrOrderAlreadyExistsForQuote = errors.New("order already exists for this quotation")
var ErrPaymentTypeInvalid = errors.New("payment type must be DP or FP")
var ErrInvalidQuotationSet = errors.New("INVALID_QUOTATION_SET")
var ErrRFQLocked = errors.New("RFQ_LOCKED")
var ErrSelfTransaction = errors.New("SELF_TRANSACTION")
var ErrPaymentAmountMismatch = errors.New("payment amount does not match order amount for payment type")
var ErrPaymentAlreadyExists = errors.New("payment already exists for this order and payment type")
var ErrPaymentStateInvalid = errors.New("payment is not in a verifiable state")
var ErrDepositAlreadyPaid = errors.New("DEPOSIT_ALREADY_PAID")
var ErrDepositExpired = errors.New("DEPOSIT_EXPIRED")
var ErrConfirmReceiptInvalidStatus = errors.New("order status must be SH")
var ErrConfirmReceiptNotAllowed = errors.New("order already completed or cancelled")
var ErrReviewRatingInvalid = errors.New("rating must be between 1 and 5")
var ErrReviewCommentInvalid = errors.New("comment must be 1-1000 characters")
var ErrReviewImagesInvalid = errors.New("image_urls must contain at most 5 unique urls")
var ErrReviewOrderNotCompleted = errors.New("order must be completed before review")
var ErrReviewAlreadyExists = errors.New("review already exists for this order")

type OrderService struct {
	db            *sqlx.DB
	repo          *orderrepo.OrderRepository
	schedules     *paymentrepo.PaymentScheduleRepository
	wallets       *walletrepo.WalletRepository
	txLedger      *walletrepo.TransactionRepository
	quotations    *quotationrepo.QuotationRepository
	rfqs          *rfqrepo.RFQRepository
	reviews       *userrepo.ReviewRepository
	notifications notificationCreator
	messages      systemMessageSender
}

func NewOrderService(db *sqlx.DB, repo *orderrepo.OrderRepository, schedules *paymentrepo.PaymentScheduleRepository, wallets *walletrepo.WalletRepository, txLedger *walletrepo.TransactionRepository, quotations *quotationrepo.QuotationRepository, rfqs *rfqrepo.RFQRepository, reviews *userrepo.ReviewRepository, notifications notificationCreator, messages systemMessageSender) *OrderService {
	return &OrderService{db: db, repo: repo, schedules: schedules, wallets: wallets, txLedger: txLedger, quotations: quotations, rfqs: rfqs, reviews: reviews, notifications: notifications, messages: messages}
}

var thailandLocation = time.FixedZone("Asia/Bangkok", 7*60*60)

func getShippingDays(db *sqlx.DB) int {
	var cfg struct {
		Value string `db:"value"`
	}
	err := db.Get(&cfg, `SELECT value FROM tconfig WHERE key = 'shipping_days'`)
	if err == nil && cfg.Value != "" {
		if n, err := strconv.Atoi(cfg.Value); err == nil && n > 0 {
			return n
		}
	}

	if s := os.Getenv("PLATFORM_SHIPPING_DAYS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return n
		}
	}

	return 7
}

func calculateEstimatedDelivery(orderDate time.Time, leadTimeDays int64, shippingDays int, deliveryDate *time.Time) time.Time {
	if deliveryDate != nil {
		return dateOnly(*deliveryDate)
	}
	est := orderDate.AddDate(0, 0, int(leadTimeDays)+shippingDays)
	return dateOnly(est)
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
