package input

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/khaopod/backend/internal/core/domain"
)

// OrderUseCase = สั่งซื้อ + ติดตาม + ยกเลิก + แจ้งชำระเงิน (driving port)
type OrderUseCase interface {
	// Place สร้างออเดอร์จากตะกร้า active ทั้งใบ — ตัดสต็อกใน tx เดียว (ดู docs/rest_api.md §9.1)
	Place(ctx context.Context, userID uint, cmd PlaceOrderCommand) (*domain.Order, error)
	// List คืนออเดอร์ของ user (ใหม่สุดก่อน) + total สำหรับ pagination
	List(ctx context.Context, userID uint, q OrderListQuery) ([]domain.Order, int, error)
	// Get คืนออเดอร์เต็ม (items + payment) — ErrNotFound ถ้าไม่ใช่ของ user นี้
	Get(ctx context.Context, userID uint, orderNumber string) (*domain.Order, error)
	// Cancel ยกเลิก + คืนสต็อกใน tx เดียว — ได้เฉพาะสถานะ pending/paid
	Cancel(ctx context.Context, userID uint, orderNumber, reason string) (*domain.Order, error)
	// SubmitPayment แจ้งโอน/แนบสลิป → payment สถานะ pending_review รอแอดมินตรวจ
	SubmitPayment(ctx context.Context, userID uint, orderNumber string, cmd SubmitPaymentCommand) (*domain.Order, error)
}

type PlaceOrderCommand struct {
	AddressID     uint
	PaymentMethod domain.PaymentMethod
	CustomerNote  string
}

type OrderListQuery struct {
	Status string // "" = ทุกสถานะ (ค่าที่ไม่รู้จักให้ inbound adapter คัดออกก่อน)
	Limit  int
	Offset int
}

// SubmitPaymentCommand — สลิปผ่านการตรวจชนิด/ขนาดจาก inbound adapter แล้ว
type SubmitPaymentCommand struct {
	Method         domain.PaymentMethod
	Amount         decimal.Decimal
	TransactionRef string
	SlipContent    []byte
	SlipExt        string // นามสกุลตามชนิดไฟล์จริง เช่น ".jpg"
}
