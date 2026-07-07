package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config รวม setting ทั้งหมดของแอป อ่านจาก environment variable เท่านั้น
// (secret ห้าม hardcode — ดู CLAUDE.md ส่วนความปลอดภัย)
type Config struct {
	Env  string // "development" | "production"
	Port string

	DatabaseURL string

	// ImageDir = โฟลเดอร์รูปสินค้าที่เสิร์ฟผ่าน static route /images
	ImageDir string

	// UploadDir = โฟลเดอร์เก็บไฟล์ที่ผู้ใช้อัปโหลด (รูปโปรไฟล์) เสิร์ฟผ่าน static route /uploads
	UploadDir string

	// CORSAllowedOrigins = รายชื่อ origin ของ frontend ที่อนุญาตให้ยิงตรงมา backend
	// (browser บังคับ CORS; ยิงตรงข้าม origin ต้องอยู่ใน allowlist นี้) — ห้ามใช้ "*" คู่กับ credentials
	CORSAllowedOrigins []string

	// JWT (access token อายุสั้น)
	JWTSecret string
	AccessTTL time.Duration

	// OTP + refresh session
	OTPSecret      string // ใช้ทำ HMAC ของ OTP ก่อนเก็บลง DB (เก็บเฉพาะแฮช)
	OTPTTL         time.Duration
	OTPLength      int
	OTPMaxAttempts int
	RefreshTTL     time.Duration

	// Rate limit ของ endpoint OTP (ปรับผ่าน env — dev ตั้งให้ recover ไว ไม่ต้องรอนาน)
	OTPReqPerIP          int
	OTPReqPerIPWindow    time.Duration
	OTPReqPerEmail       int
	OTPReqPerEmailWindow time.Duration
	OTPVerifyPerIP       int
	OTPVerifyPerIPWindow time.Duration

	// SMTP สำหรับส่ง OTP ทางอีเมล
	SMTPHost    string
	SMTPPort    int
	SMTPUser    string
	SMTPPass    string
	MailFrom    string
	SMTPTimeout time.Duration // กัน request ค้างถ้า SMTP ไม่ตอบ (dev ที่ไม่มี Mailpit)

	GoogleClientID string
}

// Load อ่าน config จาก env และตรวจว่ามี secret ที่จำเป็นครบ
func Load() (*Config, error) {
	c := &Config{
		Env:         getenv("APP_ENV", "development"),
		Port:        getenv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		ImageDir:    getenv("IMAGE_DIR", "./migrations/image"),
		UploadDir:   getenv("UPLOAD_DIR", "./uploads"),
		// dev: frontend รันที่ Vite (5173); prod ตั้ง CORS_ALLOWED_ORIGINS เป็นโดเมนจริง (คั่นด้วย comma)
		CORSAllowedOrigins: getcsv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		AccessTTL:          getdur("ACCESS_TOKEN_TTL", 15*time.Minute),
		OTPSecret:          os.Getenv("OTP_SECRET"),
		OTPTTL:             getdur("OTP_TTL", 5*time.Minute),
		OTPLength:          getint("OTP_LENGTH", 6),
		OTPMaxAttempts:     getint("OTP_MAX_ATTEMPTS", 5),
		RefreshTTL:         getdur("REFRESH_TOKEN_TTL", 30*24*time.Hour),

		// default: ขอ OTP ได้ 10 ครั้ง/10 นาที ต่อ IP, 1 ครั้ง/60วิ ต่ออีเมล, ยืนยันได้ 10 ครั้ง/นาที ต่อ IP
		OTPReqPerIP:          getint("OTP_REQ_PER_IP", 10),
		OTPReqPerIPWindow:    getdur("OTP_REQ_PER_IP_WINDOW", 10*time.Minute),
		OTPReqPerEmail:       getint("OTP_REQ_PER_EMAIL", 1),
		OTPReqPerEmailWindow: getdur("OTP_REQ_PER_EMAIL_WINDOW", time.Minute),
		OTPVerifyPerIP:       getint("OTP_VERIFY_PER_IP", 10),
		OTPVerifyPerIPWindow: getdur("OTP_VERIFY_PER_IP_WINDOW", time.Minute),

		SMTPHost:       getenv("SMTP_HOST", "localhost"),
		SMTPPort:       getint("SMTP_PORT", 1025),
		SMTPUser:       os.Getenv("SMTP_USERNAME"),
		SMTPPass:       os.Getenv("SMTP_PASSWORD"),
		MailFrom:       getenv("MAIL_FROM", "no-reply@kbcnews.shop"),
		SMTPTimeout:    getdur("SMTP_TIMEOUT", 10*time.Second),
		GoogleClientID: os.Getenv("GOOGLE_CLIENT_ID"),
	}

	var missing []string
	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if c.OTPSecret == "" {
		missing = append(missing, "OTP_SECRET")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required env: %v", missing)
	}
	return c, nil
}

func (c *Config) IsProduction() bool { return c.Env == "production" }

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getint(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// getcsv อ่านค่าแบบ comma-separated จาก env แล้วตัดช่องว่าง/ค่าว่างทิ้ง
func getcsv(key, def string) []string {
	parts := strings.Split(getenv(key, def), ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func getdur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
