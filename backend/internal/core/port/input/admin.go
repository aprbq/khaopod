package input

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/khaopod/backend/internal/core/domain"
)

// AdminOrderUseCase = หลังบ้านแอดมิน: ดูภาพรวม + จัดการออเดอร์/การชำระเงินของทุกคน (driving port)
// ทุกเมธอดถูกครอบด้วย RequireAdmin ที่ inbound adapter — core ไม่ต้องเช็ค role ซ้ำ
type AdminOrderUseCase interface {
	Summary(ctx context.Context) (*domain.AdminSummary, error)
	ListOrders(ctx context.Context, q OrderListQuery) ([]domain.Order, int, error)
	GetOrder(ctx context.Context, orderNumber string) (*domain.Order, error)
	// UpdateStatus เปลี่ยนสถานะออเดอร์ + บันทึก history (changed_by = adminID)
	// เปลี่ยนเป็น cancelled จะคืนสต็อกใน tx เดียวกัน
	UpdateStatus(ctx context.Context, adminID uint, orderNumber string, status domain.OrderStatus, note string) (*domain.Order, error)
	// VerifyPayment ตัดสินสลิปที่รอตรวจ — paid = อนุมัติ, failed = ปฏิเสธ (ให้ลูกค้าแจ้งใหม่)
	VerifyPayment(ctx context.Context, adminID uint, paymentID uint, status domain.PaymentStatus) (*domain.Order, error)
}

// AdminUserUseCase = ดูรายชื่อผู้ใช้ในระบบ (driving port)
type AdminUserUseCase interface {
	ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error)
}

// AdminCatalogUseCase = จัดการสินค้า/ตัวเลือก/รูปในหลังบ้าน (driving port)
type AdminCatalogUseCase interface {
	ListProducts(ctx context.Context, q ProductQuery) ([]domain.Product, int, error) // รวมสินค้าที่ปิดขาย
	GetProduct(ctx context.Context, id uint) (*domain.Product, error)
	CreateProduct(ctx context.Context, cmd ProductCommand) (*domain.Product, error)
	UpdateProduct(ctx context.Context, id uint, cmd ProductCommand) (*domain.Product, error)
	// DeleteProduct ลบถาวร (variants/รูป cascade) — ErrConflict ถ้ามีของอยู่ในตะกร้าลูกค้า
	DeleteProduct(ctx context.Context, id uint) error
	CreateVariant(ctx context.Context, productID uint, cmd VariantCommand) error
	UpdateVariant(ctx context.Context, variantID uint, cmd VariantCommand) error
	DeleteVariant(ctx context.Context, variantID uint) error
	// AddImage รับเนื้อไฟล์ที่ inbound adapter ตรวจชนิด/ขนาดแล้ว — รูปแรกเป็นรูปปกอัตโนมัติ
	AddImage(ctx context.Context, productID uint, content []byte, ext string) error
	DeleteImage(ctx context.Context, imageID uint) error
	SetPrimaryImage(ctx context.Context, productID, imageID uint) error
}

// ProductCommand — ข้อมูลสินค้า (สร้าง/แก้ใช้ชุดเดียวกัน)
type ProductCommand struct {
	Name        string
	Slug        string
	Description string
	BasePrice   decimal.Decimal
	CategoryID  *uint
	IsActive    bool
	IsFeatured  bool
}

// VariantCommand — ตัวเลือกสินค้า (ไซซ์/สี/ราคา/สต็อก)
type VariantCommand struct {
	Name     string
	Color    string
	SKU      string
	Price    decimal.Decimal
	Stock    int
	IsActive bool
}
