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

// ProductFilter — เงื่อนไข query ที่ service normalize แล้ว (SortColumn คัดจาก whitelist กัน SQL injection)
type ProductFilter struct {
	CategorySlug string
	Search       string
	SortColumn   string // ชื่อคอลัมน์ที่ผ่านการ whitelist แล้วเท่านั้น
	SortDesc     bool
	Limit        int
	Offset       int
}

// ProductRepository — data access ของแคตตาล็อกสินค้า (อ่านเฉพาะสินค้าที่ active)
type ProductRepository interface {
	// List คืนสินค้า (พร้อม variants/รูป/หมวดหมู่) ตาม filter + จำนวนทั้งหมดก่อน limit/offset
	List(ctx context.Context, f ProductFilter) ([]domain.Product, int, error)
	// FindBySlug คืน domain.ErrNotFound ถ้าไม่พบหรือสินค้าไม่ active
	FindBySlug(ctx context.Context, slug string) (*domain.Product, error)
	// FindVariantByID คืน variant ตัวเดียว (ไว้เช็คราคา/สต็อกตอนเพิ่มลงตะกร้า)
	// คืน domain.ErrNotFound ถ้าไม่พบหรือ variant ไม่ active
	FindVariantByID(ctx context.Context, id uint) (*domain.ProductVariant, error)
}

// CartRepository — data access ของตะกร้า (active cart ต่อผู้ใช้)
type CartRepository interface {
	// Get คืน cart active ของ user พร้อม items ที่ enrich แล้ว (ชื่อสินค้า/variant/ราคา/รูป/สต็อกปัจจุบัน)
	// ถ้ายังไม่มี cart ให้คืน &domain.Cart{UserID: userID} (ตะกร้าว่าง) ไม่ใช่ error
	Get(ctx context.Context, userID uint) (*domain.Cart, error)
	// UpsertItem บวกจำนวนเข้ารายการเดิม หรือสร้างใหม่ (merge ตาม unique(cart,variant)) — สร้าง cart ให้ถ้ายังไม่มี
	UpsertItem(ctx context.Context, userID, variantID uint, qty int) error
	// SetItemQty ตั้งจำนวนของ item — scope ด้วย userID กัน IDOR (แก้ได้เฉพาะ item ในตะกร้าตัวเอง)
	SetItemQty(ctx context.Context, userID, itemID uint, qty int) error
	// RemoveItem ลบ item — scope ด้วย userID กัน IDOR
	RemoveItem(ctx context.Context, userID, itemID uint) error
	// Clear ลบทุก item ในตะกร้า active ของ user
	Clear(ctx context.Context, userID uint) error
}
