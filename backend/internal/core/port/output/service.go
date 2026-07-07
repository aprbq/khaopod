package output

import (
	"context"

	"github.com/khaopod/backend/internal/core/domain"
)

// Mailer — ส่ง OTP ทางอีเมล (adapter mailer ไป implement)
type Mailer interface {
	SendOTP(ctx context.Context, email, code string, ttlSeconds int) error
}

// AccessClaims = ข้อมูลที่ decode ได้จาก access token
type AccessClaims struct {
	UserID   uint
	PublicID string
	Role     domain.Role
}

// Tokenizer — ออก/ตรวจ access token (JWT) — adapter auth ไป implement
type Tokenizer interface {
	IssueAccess(u *domain.User) (token string, expiresIn int, err error)
	// ParseAccess คืน domain.ErrInvalidToken ถ้า token ไม่ถูกต้อง/หมดอายุ
	ParseAccess(token string) (*AccessClaims, error)
}

// GoogleIdentity = ข้อมูลที่ได้จากการ verify Google id_token
type GoogleIdentity struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

// GoogleVerifier — verify id_token กับ Google — adapter google ไป implement
type GoogleVerifier interface {
	// Verify คืน domain.ErrGoogleToken ถ้า token ไม่ถูกต้อง
	Verify(ctx context.Context, idToken string) (*GoogleIdentity, error)
}

// TxManager = Unit of Work: ให้ core คุมขอบเขต transaction โดยไม่รู้จัก *gorm.DB
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// FileStorage — เก็บไฟล์ที่ผู้ใช้อัปโหลด (adapter storage ไป implement)
type FileStorage interface {
	// Save เขียนไฟล์ตาม relative path (เช่น "avatars/x.png") แล้วคืน public URL ที่ใช้เสิร์ฟไฟล์นั้น
	Save(ctx context.Context, relPath string, content []byte) (url string, err error)
	// Remove ลบไฟล์จาก public URL ที่ Save เคยคืน — URL ที่ไม่ใช่ของ storage นี้ (เช่นรูปจาก Google) ให้เมินเฉย ๆ
	Remove(ctx context.Context, url string) error
}
