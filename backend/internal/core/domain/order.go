package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderPaid      OrderStatus = "paid"
	OrderPreparing OrderStatus = "preparing"
	OrderShipped   OrderStatus = "shipped"
	OrderDelivered OrderStatus = "delivered"
	OrderCompleted OrderStatus = "completed"
	OrderCancelled OrderStatus = "cancelled"
	OrderRefunded  OrderStatus = "refunded"
)

type PaymentStatus string

const (
	PaymentUnpaid        PaymentStatus = "unpaid"
	PaymentPendingReview PaymentStatus = "pending_review"
	PaymentPaid          PaymentStatus = "paid"
	PaymentFailed        PaymentStatus = "failed"
	PaymentRefunded      PaymentStatus = "refunded"
)

type PaymentMethod string

const (
	MethodPromptPay    PaymentMethod = "promptpay"
	MethodBankTransfer PaymentMethod = "bank_transfer"
)

// ValidPaymentMethod — วิธีจ่ายที่ระบบรองรับตอนนี้ (cod/credit_card อยู่ใน schema เผื่ออนาคต)
func ValidPaymentMethod(m PaymentMethod) bool {
	return m == MethodPromptPay || m == MethodBankTransfer
}

// ShippingAddress = snapshot ที่อยู่จัดส่ง ณ ตอนสั่งซื้อ — ห้าม join กลับไป addresses
type ShippingAddress struct {
	Recipient   string
	Phone       string
	Address     string
	Subdistrict string
	District    string
	Province    string
	PostalCode  string
	Country     string
}

// SnapshotAddress คัดลอกที่อยู่ปัจจุบันเป็น snapshot ของออเดอร์
func SnapshotAddress(a *Address) ShippingAddress {
	return ShippingAddress{
		Recipient:   a.RecipientName,
		Phone:       a.Phone,
		Address:     a.AddressLine,
		Subdistrict: a.Subdistrict,
		District:    a.District,
		Province:    a.Province,
		PostalCode:  a.PostalCode,
		Country:     a.Country,
	}
}

// OrderItem = snapshot ชื่อ/ราคา ณ ตอนซื้อ (ราคาสินค้าเปลี่ยนทีหลังต้องไม่กระทบออเดอร์เดิม)
type OrderItem struct {
	ID          uint
	VariantID   uint
	ProductName string
	VariantName string // รวมสี เช่น "ไซซ์ M / ดำ"
	UnitPrice   decimal.Decimal
	Quantity    int
	LineTotal   decimal.Decimal
}

type Order struct {
	ID             uint
	OrderNumber    string
	UserID         uint
	UserEmail      string // เติมเฉพาะ query ฝั่งแอดมิน (list/detail หลังบ้าน)
	Subtotal       decimal.Decimal
	ShippingFee    decimal.Decimal
	DiscountAmount decimal.Decimal
	TotalAmount    decimal.Decimal
	Status         OrderStatus
	PaymentStatus  PaymentStatus
	PaymentMethod  PaymentMethod
	Items          []OrderItem
	Shipping       ShippingAddress
	CustomerNote   string
	Payment        *Payment // การแจ้งชำระเงินล่าสุด (nil = ยังไม่แจ้ง)
	PlacedAt       time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CanCancel — ยกเลิกได้เฉพาะก่อนเริ่มจัดส่ง (pending/paid) ตาม docs/rest_api.md §9.4
func (o *Order) CanCancel() bool {
	return o.Status == OrderPending || o.Status == OrderPaid
}

// CanSubmitPayment — แจ้งโอนได้เมื่อออเดอร์ยังรอจ่าย (unpaid หรือแจ้งครั้งก่อนถูกปฏิเสธ)
func (o *Order) CanSubmitPayment() bool {
	if o.Status != OrderPending {
		return false
	}
	return o.PaymentStatus == PaymentUnpaid || o.PaymentStatus == PaymentFailed
}

// AdminStatuses — สถานะที่แอดมินตั้งได้ผ่านหลังบ้าน (pending เกิดจากระบบเท่านั้น)
func ValidAdminStatus(s OrderStatus) bool {
	switch s {
	case OrderPaid, OrderPreparing, OrderShipped, OrderDelivered, OrderCompleted, OrderCancelled, OrderRefunded:
		return true
	}
	return false
}

// AdminSummary = ตัวเลขสรุปหน้าแดชบอร์ดหลังบ้าน
type AdminSummary struct {
	OrdersTotal           int
	OrdersPending         int             // รอชำระ/รอจัดการ
	PaymentsPendingReview int             // สลิปที่รอตรวจ
	RevenuePaid           decimal.Decimal // ยอดขายรวมเฉพาะออเดอร์ที่จ่ายแล้ว
}

// Payment = การแจ้งชำระเงินหนึ่งครั้ง (อัปสลิปรอแอดมินตรวจ)
type Payment struct {
	ID             uint
	OrderID        uint
	Method         PaymentMethod
	Amount         decimal.Decimal
	Status         PaymentStatus
	SlipURL        string
	TransactionRef string
	PaidAt         *time.Time
	CreatedAt      time.Time
}
