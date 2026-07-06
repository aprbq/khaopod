package rest

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/khaopod/backend/internal/adapter/inbound/rest/response"
	"github.com/khaopod/backend/internal/core/port/output"
)

// RateLimitConfig = โควตา + หน้าต่างเวลาของ endpoint OTP (มาจาก config/env)
type RateLimitConfig struct {
	OTPRequestPerIP          int
	OTPRequestPerIPWindow    time.Duration
	OTPRequestPerEmail       int
	OTPRequestPerEmailWindow time.Duration
	OTPVerifyPerIP           int
	OTPVerifyPerIPWindow     time.Duration
}

// withDefaults เติมค่าปลอดภัยถ้ามีฟิลด์ไหนไม่ถูกตั้ง (กันลืม wire แล้วกลายเป็น 0 = บล็อกทุกครั้ง)
func (r RateLimitConfig) withDefaults() RateLimitConfig {
	if r.OTPRequestPerIP <= 0 {
		r.OTPRequestPerIP = 10
	}
	if r.OTPRequestPerIPWindow <= 0 {
		r.OTPRequestPerIPWindow = 10 * time.Minute
	}
	if r.OTPRequestPerEmail <= 0 {
		r.OTPRequestPerEmail = 1
	}
	if r.OTPRequestPerEmailWindow <= 0 {
		r.OTPRequestPerEmailWindow = time.Minute
	}
	if r.OTPVerifyPerIP <= 0 {
		r.OTPVerifyPerIP = 10
	}
	if r.OTPVerifyPerIPWindow <= 0 {
		r.OTPVerifyPerIPWindow = time.Minute
	}
	return r
}

// Deps = สิ่งที่ router ต้องใช้ (ประกอบมาจาก composition root)
type Deps struct {
	Auth        *AuthHandler
	User        *UserHandler
	Product     *ProductHandler
	Cart        *CartHandler
	Tokens      output.Tokenizer
	RateLimit   RateLimitConfig
	ImageDir    string   // โฟลเดอร์รูปสินค้าที่เสิร์ฟผ่าน /images
	CORSOrigins []string // origin ของ frontend ที่อนุญาตให้ยิงข้าม origin ได้
}

// RegisterRoutes ผูก path ทั้งหมด — path/method ต้องตรงกับ docs/rest_api.md
func RegisterRoutes(engine *gin.Engine, d Deps) {
	rl := d.RateLimit.withDefaults()

	// CORS ต้องมาก่อนทุก route (รวม /images, /healthz) เพื่อจับ preflight ให้ครบ
	engine.Use(CORS(d.CORSOrigins))

	// rate limiter สำหรับ endpoint OTP (กัน brute-force / spam)
	otpPerIP := NewLimiter(rl.OTPRequestPerIP, rl.OTPRequestPerIPWindow)
	otpPerEmail := NewLimiter(rl.OTPRequestPerEmail, rl.OTPRequestPerEmailWindow)
	verifyPerIP := NewLimiter(rl.OTPVerifyPerIP, rl.OTPVerifyPerIPWindow)

	engine.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})

	// เสิร์ฟรูปสินค้าแบบ static (🔓) — product_images.url ชี้มาที่ /images/<ไฟล์>
	if d.ImageDir != "" {
		engine.Static("/images", d.ImageDir)
	}

	v1 := engine.Group("/api/v1")

	// ---- Auth (🔓) ----
	v1.POST("/auth/otp/request", OTPRequestRateLimit(otpPerIP, otpPerEmail), d.Auth.RequestOTP)
	v1.POST("/auth/google", OTPRequestRateLimit(otpPerIP, otpPerEmail), d.Auth.Google)
	v1.POST("/auth/otp/verify", IPRateLimit(verifyPerIP), d.Auth.VerifyOTP)
	v1.POST("/auth/refresh", d.Auth.Refresh)

	// ---- Products (🔓 สาธารณะ) ----
	v1.GET("/products", d.Product.List)
	v1.GET("/products/:slug", d.Product.GetBySlug)

	// ---- ต้องล็อกอิน (🔒) ----
	secured := v1.Group("", RequireAuth(d.Tokens))
	secured.POST("/auth/logout", d.Auth.Logout)
	secured.GET("/auth/me", d.Auth.Me)
	secured.GET("/me", d.User.GetMe)
	secured.PATCH("/me", d.User.UpdateMe)

	// ---- Cart (🔒) — ดู docs/rest_api.md §6 ----
	secured.GET("/cart", d.Cart.Get)
	secured.POST("/cart/items", d.Cart.AddItem)
	secured.PATCH("/cart/items/:itemId", d.Cart.UpdateItem)
	secured.DELETE("/cart/items/:itemId", d.Cart.RemoveItem)
	secured.DELETE("/cart", d.Cart.Clear)
}
