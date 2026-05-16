package api

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/wemake/internal/config"
	"github.com/yourusername/wemake/internal/handler"
	"github.com/yourusername/wemake/internal/media"
	"github.com/yourusername/wemake/internal/repository"
	"github.com/yourusername/wemake/internal/service"
)

type routeHandlers struct {
	authService *service.AuthService

	auth              *handler.AuthHandler
	catalog           *handler.CatalogHandler
	address           *handler.AddressHandler
	wallet            *handler.WalletHandler
	rfq               *handler.RFQHandler
	quotation         *handler.QuotationHandler
	order             *handler.OrderHandler
	orderPayment      *handler.OrderPaymentHandler
	production        *handler.ProductionHandler
	message           *handler.MessageHandler
	master            *handler.MasterHandler
	transaction       *handler.TransactionHandler
	frontend          *handler.FrontendHandler
	media             *handler.MediaHandler
	review            *handler.ReviewHandler
	conversation      *handler.ConversationHandler
	notification      *handler.NotificationHandler
	showcase          *handler.ShowcaseHandler
	boq               *handler.BOQHandler
	profile           *handler.ProfileHandler
	factory           *handler.FactoryHandler
	favorite          *handler.FavoriteHandler
	certificate       *handler.CertificateHandler
	settlement        *handler.SettlementHandler
	topup             *handler.TopupHandler
	withdrawal        *handler.WithdrawalHandler
	dispute           *handler.DisputeHandler
	quotationTemplate *handler.QuotationTemplateHandler
	paymentSchedule   *handler.PaymentScheduleHandler
	platformConfig    *handler.PlatformConfigHandler
	adminFactory      *handler.AdminFactoryHandler
	adminDashboard    *handler.AdminDashboardHandler
	adminRFQ          *handler.AdminRFQHandler
	adminOrder        *handler.AdminOrderHandler
	adminConfig       *handler.AdminConfigHandler
	adminUser         *handler.AdminUserHandler
	adminCustomer     *handler.AdminCustomerHandler
}

func newRouteHandlers(db *sqlx.DB, cfg *config.Config) *routeHandlers {
	authRepo := repository.NewAuthRepository(db)
	catalogRepo := repository.NewCatalogRepository(db)
	addressRepo := repository.NewAddressRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	rfqRepo := repository.NewRFQRepository(db)
	quotationRepo := repository.NewQuotationRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	productionRepo := repository.NewProductionRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	masterRepo := repository.NewMasterRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	frontendRepo := repository.NewFrontendRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	conversationRepo := repository.NewConversationRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	rfqItemRepo := repository.NewRFQItemRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	showcaseRepo := repository.NewShowcaseRepository(db)
	factoryRepo := repository.NewFactoryRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	certificateRepo := repository.NewCertificateRepository(db)
	settlementRepo := repository.NewSettlementRepository(db)
	topupRepo := repository.NewTopupRepository(db)
	withdrawalRepo := repository.NewWithdrawalRepository(db)
	disputeRepo := repository.NewDisputeRepository(db)
	quotationTemplateRepo := repository.NewQuotationTemplateRepository(db)
	paymentScheduleRepo := repository.NewPaymentScheduleRepository(db)
	platformConfigRepo := repository.NewPlatformConfigRepository(db)
	quotationItemRepo := repository.NewQuotationItemRepository(db)
	commissionRepo := repository.NewCommissionRepository(db)
	adminAuditRepo := repository.NewAdminAuditRepository(db)
	adminDashboardRepo := repository.NewAdminDashboardRepository(db)
	customerAdminRepo := repository.NewCustomerAdminRepository(db)
	settlementAdminRepo := repository.NewSettlementAdminRepository(db)

	authService := service.NewAuthService(authRepo, cfg.JWTSecret)
	catalogService := service.NewCatalogService(catalogRepo)
	addressService := service.NewAddressService(addressRepo, factoryRepo)
	walletService := service.NewWalletService(walletRepo, transactionRepo)
	notificationService := service.NewNotificationService(notificationRepo)
	rfqService := service.NewRFQService(rfqRepo, factoryRepo, notificationService)
	messageService := service.NewMessageService(messageRepo, conversationRepo, notificationService)
	orderService := service.NewOrderService(db, orderRepo, paymentScheduleRepo, walletRepo, transactionRepo, quotationRepo, rfqRepo, reviewRepo, notificationService, messageService)
	commissionService := service.NewCommissionService(platformConfigRepo, commissionRepo)
	platformConfigService := service.NewPlatformConfigService(db, platformConfigRepo, adminAuditRepo)
	quotationService := service.NewQuotationService(db, quotationRepo, rfqRepo, quotationItemRepo, commissionService, orderService, factoryRepo, notificationService, messageService)
	orderPaymentService := service.NewOrderPaymentService(db)
	productionService := service.NewProductionService(productionRepo)
	masterService := service.NewMasterService(masterRepo)
	transactionService := service.NewTransactionService(transactionRepo)
	frontendService := service.NewFrontendService(frontendRepo, factoryRepo)
	reviewService := service.NewReviewService(reviewRepo)
	conversationService := service.NewConversationService(conversationRepo, rfqRepo, messageService)
	showcaseService := service.NewShowcaseService(showcaseRepo, factoryRepo)
	boqService := service.NewBOQService(db, conversationRepo, rfqRepo, rfqItemRepo, quotationRepo, quotationItemRepo, orderService, messageService, notificationService, commissionService)
	profileService := service.NewProfileService(profileRepo, authRepo)
	factoryService := service.NewFactoryService(factoryRepo)
	favoriteService := service.NewFavoriteService(favoriteRepo)
	certificateService := service.NewCertificateService(certificateRepo)
	settlementService := service.NewSettlementService(settlementRepo)
	topupService := service.NewTopupService(topupRepo, walletRepo)
	withdrawalService := service.NewWithdrawalService(withdrawalRepo, walletRepo)
	disputeService := service.NewDisputeService(disputeRepo)
	quotationTemplateService := service.NewQuotationTemplateService(quotationTemplateRepo)
	paymentScheduleService := service.NewPaymentScheduleService(paymentScheduleRepo)
	adminFactoryService := service.NewAdminFactoryService(factoryRepo, adminAuditRepo, commissionRepo, platformConfigRepo)
	adminDashboardService := service.NewAdminDashboardService(adminDashboardRepo)

	cld, err := media.NewCloudinaryClient(cfg)
	if err != nil {
		log.Printf("cloudinary disabled: invalid configuration: %v", err)
		cld = nil
	}

	return &routeHandlers{
		authService:       authService,
		auth:              handler.NewAuthHandler(authService),
		catalog:           handler.NewCatalogHandler(catalogService),
		address:           handler.NewAddressHandler(addressService),
		wallet:            handler.NewWalletHandler(walletService),
		rfq:               handler.NewRFQHandler(rfqService, authService),
		quotation:         handler.NewQuotationHandler(quotationService, authService),
		order:             handler.NewOrderHandler(orderService, authService),
		orderPayment:      handler.NewOrderPaymentHandler(orderPaymentService),
		production:        handler.NewProductionHandler(productionService),
		message:           handler.NewMessageHandler(messageService),
		master:            handler.NewMasterHandler(masterService),
		transaction:       handler.NewTransactionHandler(transactionService),
		frontend:          handler.NewFrontendHandler(frontendService),
		media:             handler.NewMediaHandler(cfg.PublicBaseURL, cld),
		review:            handler.NewReviewHandler(reviewService),
		conversation:      handler.NewConversationHandler(conversationService),
		notification:      handler.NewNotificationHandler(notificationService),
		showcase:          handler.NewShowcaseHandler(showcaseService),
		boq:               handler.NewBOQHandler(boqService),
		profile:           handler.NewProfileHandler(profileService, cfg.PublicBaseURL, cld),
		factory:           handler.NewFactoryHandler(factoryService, authService),
		favorite:          handler.NewFavoriteHandler(favoriteService),
		certificate:       handler.NewCertificateHandler(certificateService),
		settlement:        handler.NewSettlementHandler(settlementService),
		topup:             handler.NewTopupHandler(topupService),
		withdrawal:        handler.NewWithdrawalHandler(withdrawalService),
		dispute:           handler.NewDisputeHandler(disputeService),
		quotationTemplate: handler.NewQuotationTemplateHandler(quotationTemplateService),
		paymentSchedule:   handler.NewPaymentScheduleHandler(paymentScheduleService),
		platformConfig:    handler.NewPlatformConfigHandler(platformConfigService, authService),
		adminFactory:      handler.NewAdminFactoryHandler(factoryRepo, adminFactoryService),
		adminDashboard:    handler.NewAdminDashboardHandler(adminDashboardService),
		adminRFQ:          handler.NewAdminRFQHandler(rfqRepo, adminAuditRepo),
		adminOrder:        handler.NewAdminOrderHandler(orderRepo, orderService, withdrawalRepo, disputeRepo, adminAuditRepo),
		adminConfig:       handler.NewAdminConfigHandler(commissionRepo, adminAuditRepo),
		adminUser:         handler.NewAdminUserHandler(authService, authRepo),
		adminCustomer:     handler.NewAdminCustomerHandler(customerAdminRepo, settlementAdminRepo),
	}
}
