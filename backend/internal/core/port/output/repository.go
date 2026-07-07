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
	// ListAll คืนผู้ใช้ทุกคน (ใหม่สุดก่อน) + จำนวนทั้งหมด — ใช้ในหลังบ้านแอดมิน
	ListAll(ctx context.Context, limit, offset int) ([]domain.User, int, error)
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
	// GetVariantForUpdate ล็อกแถว variant (SELECT ... FOR UPDATE) — เรียกได้เฉพาะใน WithinTx
	// ใช้ตอนตัด/คืนสต็อกเพื่อกัน oversell; คืน domain.ErrNotFound ถ้าไม่พบ
	GetVariantForUpdate(ctx context.Context, id uint) (*domain.ProductVariant, error)
	// SaveVariantStock เขียน stock_quantity ของ variant ที่ล็อกไว้กลับลง DB
	SaveVariantStock(ctx context.Context, v *domain.ProductVariant) error
	// ListCategories คืนหมวดหมู่ที่ active เรียงตาม sort_order
	ListCategories(ctx context.Context) ([]domain.Category, error)
}

// CatalogAdminRepository — เขียนแคตตาล็อกจากหลังบ้าน (implement โดย postgres adapter ตัวเดียวกับ ProductRepo)
type CatalogAdminRepository interface {
	// ListAllProducts เหมือน ProductRepository.List แต่รวมสินค้าที่ปิดขาย (is_active = false)
	ListAllProducts(ctx context.Context, f ProductFilter) ([]domain.Product, int, error)
	// FindProductByID คืนสินค้า (รวม inactive) พร้อม variants + รูปทั้งหมด
	FindProductByID(ctx context.Context, id uint) (*domain.Product, error)
	CreateProduct(ctx context.Context, p *domain.Product) error // เซ็ต ID กลับให้
	UpdateProduct(ctx context.Context, p *domain.Product) error
	// DeleteProduct ลบถาวร — คืน domain.ErrConflict ถ้า FK ขวาง (เช่น variant อยู่ในตะกร้าลูกค้า)
	DeleteProduct(ctx context.Context, id uint) error
	CreateVariant(ctx context.Context, v *domain.ProductVariant) error
	UpdateVariant(ctx context.Context, v *domain.ProductVariant) error
	DeleteVariant(ctx context.Context, id uint) error // ErrConflict ถ้า FK ขวาง
	AddImage(ctx context.Context, img *domain.ProductImage) error
	// DeleteImage ลบแถวรูปแล้วคืน URL เดิม (ให้ service ลบไฟล์ต่อแบบ best-effort)
	DeleteImage(ctx context.Context, id uint) (url string, err error)
	// SetPrimaryImage ปลดรูปปกเดิมแล้วตั้งรูปใหม่ — เรียกใน WithinTx (DB มี unique partial index คุม)
	SetPrimaryImage(ctx context.Context, productID, imageID uint) error
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
	// MarkConverted เปลี่ยนตะกร้าเป็น converted หลังสร้างออเดอร์สำเร็จ (เรียกใน tx เดียวกับ checkout)
	MarkConverted(ctx context.Context, cartID uint) error
}

// AddressRepository — data access ของที่อยู่จัดส่ง (ทุก query scope ด้วย userID กัน IDOR)
type AddressRepository interface {
	ListByUser(ctx context.Context, userID uint) ([]domain.Address, error)
	// FindByID คืน domain.ErrNotFound ถ้าไม่พบหรือไม่ใช่ของ user นี้
	FindByID(ctx context.Context, userID, id uint) (*domain.Address, error)
	Create(ctx context.Context, a *domain.Address) error // เซ็ต ID กลับให้หลัง insert
	Update(ctx context.Context, a *domain.Address) error
	Delete(ctx context.Context, userID, id uint) error
	// ClearDefault ปลด default ของทุกที่อยู่ของ user (เรียกก่อนตั้ง default ใหม่ — DB มี unique partial index คุมอยู่)
	ClearDefault(ctx context.Context, userID uint) error
}

// OrderRepository — data access ของคำสั่งซื้อ + การชำระเงิน
type OrderRepository interface {
	// Create insert order + items + history แถวแรก (สถานะ pending) — DB เป็นคน gen order_number
	// แล้วเซ็ต ID/OrderNumber/PlacedAt กลับเข้า struct
	Create(ctx context.Context, o *domain.Order) error
	// ListByUser คืนออเดอร์ (พร้อม items) ใหม่สุดก่อน + จำนวนทั้งหมดก่อน limit/offset
	ListByUser(ctx context.Context, userID uint, q OrderListFilter) ([]domain.Order, int, error)
	// FindByNumber คืนออเดอร์เต็ม (items + payment ล่าสุด) — scope ด้วย userID, ErrNotFound ถ้าไม่ใช่ของ user
	FindByNumber(ctx context.Context, userID uint, orderNumber string) (*domain.Order, error)
	// UpdateStatus เขียนสถานะ order/payment_status + เพิ่มแถว history (changedBy 0 = ระบบ/ลูกค้า)
	UpdateStatus(ctx context.Context, orderID uint, status domain.OrderStatus, paymentStatus domain.PaymentStatus, note string, changedBy uint) error
	// CreatePayment insert การแจ้งชำระเงินหนึ่งครั้ง
	CreatePayment(ctx context.Context, p *domain.Payment) error

	// ---- ฝั่งแอดมิน (ไม่ scope ด้วย userID — ครอบสิทธิ์ที่ inbound middleware แล้ว) ----

	// ListAll คืนออเดอร์ทุกคน (พร้อม items + UserEmail) ใหม่สุดก่อน + total
	ListAll(ctx context.Context, f OrderListFilter) ([]domain.Order, int, error)
	// FindByNumberAny เหมือน FindByNumber แต่ไม่จำกัดเจ้าของ (เติม UserEmail ให้ด้วย)
	FindByNumberAny(ctx context.Context, orderNumber string) (*domain.Order, error)
	// FindByIDAny คืนออเดอร์ตาม id ไม่จำกัดเจ้าของ (ใช้ตามรอยจาก payment.OrderID)
	FindByIDAny(ctx context.Context, id uint) (*domain.Order, error)
	// FindPaymentByID คืน payment ตาม id — ErrNotFound ถ้าไม่พบ
	FindPaymentByID(ctx context.Context, id uint) (*domain.Payment, error)
	// SavePaymentVerdict เขียนผลตรวจสลิป (status + paid_at + verified_by)
	SavePaymentVerdict(ctx context.Context, paymentID uint, status domain.PaymentStatus, verifiedBy uint) error
	// Summary รวมตัวเลขหน้าแดชบอร์ด
	Summary(ctx context.Context) (*domain.AdminSummary, error)
}

// OrderListFilter — เงื่อนไข list ที่ service ตรวจแล้ว (Status ว่าง = ทุกสถานะ)
type OrderListFilter struct {
	Status domain.OrderStatus
	Limit  int
	Offset int
}
