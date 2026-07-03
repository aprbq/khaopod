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
	txMgr := postgres.NewTxManager(db)
	mail := mailer.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.MailFrom)
	tokenizer := auth.NewJWTTokenizer(cfg.JWTSecret, cfg.AccessTTL)
	googleVerifier := google.NewVerifier(cfg.GoogleClientID)

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
	userSvc := service.NewUserService(userRepo)

	// inbound adapter (driving side)
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.Default()
	rest.RegisterRoutes(engine, rest.Deps{
		Auth:   rest.NewAuthHandler(authSvc, userSvc, int(cfg.RefreshTTL.Seconds()), cfg.IsProduction()),
		User:   rest.NewUserHandler(userSvc),
		Tokens: tokenizer,
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
