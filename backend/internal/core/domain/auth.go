package domain

import (
	"crypto/subtle"
	"time"
)

type OTPPurpose string

const (
	OTPPurposeLogin       OTPPurpose = "login"
	OTPPurposeVerifyEmail OTPPurpose = "verify_email"
	OTPPurposeChangeEmail OTPPurpose = "change_email"
)

// OTPCode = รหัส OTP ที่ส่งไปทางอีเมล — เก็บเฉพาะ "แฮช" ของโค้ด ไม่เก็บเลขจริง
type OTPCode struct {
	ID          uint
	UserID      *uint // อาจ nil ถ้ายังไม่มี user ตอนขอ OTP
	Email       string
	Purpose     OTPPurpose
	CodeHash    string
	ExpiresAt   time.Time
	ConsumedAt  *time.Time
	Attempts    int
	MaxAttempts int
	RequestIP   string
	CreatedAt   time.Time
}

func (o *OTPCode) IsConsumed() bool           { return o.ConsumedAt != nil }
func (o *OTPCode) IsLocked() bool             { return o.Attempts >= o.MaxAttempts }
func (o *OTPCode) IsExpired(t time.Time) bool { return t.After(o.ExpiresAt) }

// Verify เทียบแฮชของโค้ดที่ผู้ใช้กรอก (providedHash) กับที่เก็บไว้
// - สำเร็จ: mark consumed แล้วคืน nil
// - ผิด: เพิ่ม Attempts แล้วคืน ErrInvalidOTP (หรือ ErrTooManyAttempts ถ้าครบเพดาน)
// เรียกฝั่ง service ต้อง persist state (Attempts/ConsumedAt) ที่เปลี่ยนไปเสมอ
func (o *OTPCode) Verify(providedHash string, now time.Time) error {
	switch {
	case o.IsConsumed():
		return ErrInvalidOTP
	case o.IsLocked():
		return ErrTooManyAttempts
	case o.IsExpired(now):
		return ErrOTPExpired
	}

	// เทียบแบบ constant-time กัน timing attack
	if subtle.ConstantTimeCompare([]byte(o.CodeHash), []byte(providedHash)) != 1 {
		o.Attempts++
		if o.IsLocked() {
			return ErrTooManyAttempts
		}
		return ErrInvalidOTP
	}

	consumed := now
	o.ConsumedAt = &consumed
	return nil
}

// Session = refresh token ที่ออกหลังล็อกอินสำเร็จ (เก็บเฉพาะแฮชของ token)
type Session struct {
	ID               uint
	UserID           uint
	RefreshTokenHash string
	UserAgent        string
	IPAddress        string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
}

func (s *Session) IsRevoked() bool            { return s.RevokedAt != nil }
func (s *Session) IsExpired(t time.Time) bool { return t.After(s.ExpiresAt) }
func (s *Session) IsValid(t time.Time) bool   { return !s.IsRevoked() && !s.IsExpired(t) }

func (s *Session) Revoke(now time.Time) {
	if s.RevokedAt == nil {
		s.RevokedAt = &now
	}
}

// OAuthAccount = บัญชี OAuth (Google) ที่ผูกกับ user
type OAuthAccount struct {
	ID             uint
	UserID         uint
	Provider       string // "google"
	ProviderUserID string // Google "sub"
	ProviderEmail  string
	CreatedAt      time.Time
}
