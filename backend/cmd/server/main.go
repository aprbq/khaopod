package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/khaopod/backend/internal/adapter/inbound/rest"
	"github.com/khaopod/backend/internal/adapter/outbound/auth"
	"github.com/khaopod/backend/internal/adapter/outbound/google"
	"github.com/khaopod/backend/internal/adapter/outbound/mailer"
	"github.com/khaopod/backend/internal/adapter/outbound/postgres"
	"github.com/khaopod/backend/internal/adapter/outbound/storage"
	"github.com/khaopod/backend/internal/config"
	"github.com/khaopod/backend/internal/core/service"
)

// composition root — ที่เดียวที่ประกอบ adapter → service → handler เข้าด้วยกัน
func main() {
	// โหลด .env ตอน dev (best-effort) — prod ที่ตั้ง env เองไม่มีไฟล์นี้ก็ข้ามไป
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := postgres.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	// outbound adapters (driven side)
	userRepo := postgres.NewUserRepo(db)
	otpRepo := postgres.NewOTPRepo(db)
	sessionRepo := postgres.NewSessionRepo(db)
	oauthRepo := postgres.NewOAuthRepo(db)
	productRepo := postgres.NewProductRepo(db)
	cartRepo := postgres.NewCartRepo(db)
	addressRepo := postgres.NewAddressRepo(db)
	orderRepo := postgres.NewOrderRepo(db)
	txMgr := postgres.NewTxManager(db)
	mail := mailer.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.MailFrom, cfg.SMTPTimeout)
	tokenizer := auth.NewJWTTokenizer(cfg.JWTSecret, cfg.AccessTTL)
	googleVerifier := google.NewVerifier(cfg.GoogleClientID)
	fileStore, err := storage.NewLocal(cfg.UploadDir, "/uploads")
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	// core services (use cases) — inject output ports
	authSvc := service.NewAuthService(
		userRepo, otpRepo, sessionRepo, oauthRepo, mail, tokenizer, googleVerifier, txMgr,
		service.AuthConfig{
			OTPSecret:      cfg.OTPSecret,
			OTPTTL:         cfg.OTPTTL,
			OTPLength:      cfg.OTPLength,
			OTPMaxAttempts: cfg.OTPMaxAttempts,
			RefreshTTL:     cfg.RefreshTTL,
		},
	)
	userSvc := service.NewUserService(userRepo, fileStore)
	productSvc := service.NewProductService(productRepo)
	cartSvc := service.NewCartService(cartRepo, productRepo)
	addressSvc := service.NewAddressService(addressRepo, txMgr)
	orderSvc := service.NewOrderService(orderRepo, cartRepo, productRepo, addressRepo, fileStore, txMgr)
	adminSvc := service.NewAdminOrderService(orderRepo, productRepo, txMgr)
	adminCatalogSvc := service.NewAdminCatalogService(productRepo, productRepo, fileStore, txMgr)
	adminUserSvc := service.NewAdminUserService(userRepo)

	// inbound adapter (driving side)
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.Default()
	rest.RegisterRoutes(engine, rest.Deps{
		Auth:        rest.NewAuthHandler(authSvc, userSvc, int(cfg.RefreshTTL.Seconds()), cfg.IsProduction()),
		User:        rest.NewUserHandler(userSvc),
		Product:     rest.NewProductHandler(productSvc),
		Cart:        rest.NewCartHandler(cartSvc),
		Address:     rest.NewAddressHandler(addressSvc),
		Order:       rest.NewOrderHandler(orderSvc),
		Admin:       rest.NewAdminHandler(adminSvc, adminCatalogSvc, adminUserSvc),
		Tokens:      tokenizer,
		ImageDir:    cfg.ImageDir,
		UploadDir:   cfg.UploadDir,
		CORSOrigins: cfg.CORSAllowedOrigins,
		RateLimit: rest.RateLimitConfig{
			OTPRequestPerIP:          cfg.OTPReqPerIP,
			OTPRequestPerIPWindow:    cfg.OTPReqPerIPWindow,
			OTPRequestPerEmail:       cfg.OTPReqPerEmail,
			OTPRequestPerEmailWindow: cfg.OTPReqPerEmailWindow,
			OTPVerifyPerIP:           cfg.OTPVerifyPerIP,
			OTPVerifyPerIPWindow:     cfg.OTPVerifyPerIPWindow,
		},
	})

	log.Printf("listening on :%s (env=%s)", cfg.Port, cfg.Env)
	if err := engine.Run("localhost:" + cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
