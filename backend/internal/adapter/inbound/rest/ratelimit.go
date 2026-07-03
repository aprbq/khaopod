package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/khaopod/backend/internal/adapter/inbound/rest/response"
	"github.com/khaopod/backend/internal/core/domain"
)

// Limiter = fixed-window rate limiter แบบง่ายเก็บใน memory
// (เพียงพอสำหรับ single instance — ถ้า scale หลาย instance ค่อยเปลี่ยนไปใช้ Redis)
type Limiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	hits   map[string]*window
}

type window struct {
	count   int
	resetAt time.Time
}

func NewLimiter(max int, w time.Duration) *Limiter {
	return &Limiter{max: max, window: w, hits: map[string]*window{}}
}

// Allow คืน true ถ้ายังไม่เกินโควตาในหน้าต่างเวลานี้ (นับ hit ไปด้วย)
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	w, ok := l.hits[key]
	if !ok || now.After(w.resetAt) {
		l.hits[key] = &window{count: 1, resetAt: now.Add(l.window)}
		return true
	}
	if w.count >= l.max {
		return false
	}
	w.count++
	return true
}

// OTPRequestRateLimit จำกัดทั้งต่อ IP และต่ออีเมล เพื่อกัน brute-force / spam OTP
func OTPRequestRateLimit(perIP, perEmail *Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !perIP.Allow("ip:" + c.ClientIP()) {
			response.Error(c, 429, response.CodeRateLimited, "ขอรหัสถี่เกินไป กรุณาลองใหม่ภายหลัง")
			return
		}
		if email := peekEmail(c); email != "" {
			if !perEmail.Allow("email:" + domain.NormalizeEmail(email)) {
				response.Error(c, 429, response.CodeRateLimited, "ขอรหัสถี่เกินไปสำหรับอีเมลนี้ กรุณารอสักครู่")
				return
			}
		}
		c.Next()
	}
}

// IPRateLimit จำกัดต่อ IP อย่างเดียว (ใช้กับ verify OTP กัน brute-force)
func IPRateLimit(perIP *Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !perIP.Allow("ip:" + c.ClientIP()) {
			response.Error(c, 429, response.CodeRateLimited, "ลองถี่เกินไป กรุณาลองใหม่ภายหลัง")
			return
		}
		c.Next()
	}
}

// peekEmail อ่านฟิลด์ email จาก body โดยไม่ทำให้ handler อ่าน body ต่อไม่ได้
func peekEmail(c *gin.Context) string {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body)) // คืน body ให้ handler
	var p struct {
		Email string `json:"email"`
	}
	_ = json.Unmarshal(body, &p)
	return p.Email
}
