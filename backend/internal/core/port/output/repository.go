package output

import (
	"context"

	"github.com/khaopod/backend/internal/core/domain"
)

// UserRepository — data access ของ user (adapter postgres ไป implement)
type UserRepository interface {
	FindByID(ctx context.Context, id uint) (*domain.User, error)
	FindByPublicID(ctx context.Context, publicID string) (*domain.User, error)
	// FindByEmail คืน domain.ErrNotFound ถ้าไม่พบ
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, u *domain.User) error // เซ็ต ID กลับให้หลัง insert
	Update(ctx context.Context, u *domain.User) error
}

// OTPRepository — จัดเก็บ OTP (เก็บเฉพาะแฮช)
type OTPRepository interface {
	Create(ctx context.Context, o *domain.OTPCode) error
	// FindLatestActive คืน OTP ล่าสุดที่ยังไม่ถูกใช้ (consumed_at IS NULL) สำหรับ email+purpose
	// ไม่กรองหมดอายุ/ล็อก เพื่อให้ domain แยกแยะ error ได้ (หมดอายุ vs ผิด vs ล็อก)
	FindLatestActive(ctx context.Context, email string, purpose domain.OTPPurpose) (*domain.OTPCode, error)
	// Save อัปเดต attempts / consumed_at ของ OTP ที่มีอยู่
	Save(ctx context.Context, o *domain.OTPCode) error
	// InvalidateActive mark OTP ที่ยัง active ทั้งหมดของ email+purpose ว่าใช้แล้ว (ออก OTP ใหม่ทับ)
	InvalidateActive(ctx context.Context, email string, purpose domain.OTPPurpose) error
}

// SessionRepository — refresh token session
type SessionRepository interface {
	Create(ctx context.Context, s *domain.Session) error
	// FindByTokenHash คืน domain.ErrNotFound ถ้าไม่พบ
	FindByTokenHash(ctx context.Context, hash string) (*domain.Session, error)
	Save(ctx context.Context, s *domain.Session) error
}

// OAuthRepository — บัญชี OAuth (Google) ที่ผูกกับ user
type OAuthRepository interface {
	// Upsert สร้างหรืออัปเดต link ตาม (provider, provider_user_id)
	Upsert(ctx context.Context, a *domain.OAuthAccount) error
}
