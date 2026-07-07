package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
	"github.com/khaopod/backend/internal/core/port/output"
)

// ค่าส่งเหมาจ่ายทั้งออเดอร์ — ตรงกับตัวอย่างใน docs/rest_api.md §9.1 (ยังไม่มีสูตรตามน้ำหนัก/โซน)
var flatShippingFee = decimal.NewFromInt(40)

// OrderService = use case คำสั่งซื้อ + การชำระเงิน (implements input.OrderUseCase)
type OrderService struct {
	orders   output.OrderRepository
	carts    output.CartRepository
	products output.ProductRepository
	addrs    output.AddressRepository
	files    output.FileStorage
	tx       output.TxManager
}

var _ input.OrderUseCase = (*OrderService)(nil)

func NewOrderService(
	orders output.OrderRepository,
	carts output.CartRepository,
	products output.ProductRepository,
	addrs output.AddressRepository,
	files output.FileStorage,
	tx output.TxManager,
) *OrderService {
	return &OrderService{orders: orders, carts: carts, products: products, addrs: addrs, files: files, tx: tx}
}

// Place: ตรวจสต็อก → ตัดสต็อก (row lock) → snapshot ที่อยู่+ราคา → สร้าง order → ปิดตะกร้า
// ทั้งหมดใน tx เดียว — พังตรงไหน rollback หมด สต็อกไม่รั่ว
func (s *OrderService) Place(ctx context.Context, userID uint, cmd input.PlaceOrderCommand) (*domain.Order, error) {
	if !domain.ValidPaymentMethod(cmd.PaymentMethod) {
		return nil, domain.ErrInvalidInput
	}

	var order *domain.Order
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		cart, err := s.carts.Get(ctx, userID)
		if err != nil {
			return err
		}
		if len(cart.Items) == 0 {
			return domain.ErrCartEmpty
		}

		addr, err := s.addrs.FindByID(ctx, userID, cmd.AddressID)
		if err != nil {
			return err
		}

		items := make([]domain.OrderItem, 0, len(cart.Items))
		subtotal := decimal.Zero
		for _, it := range cart.Items {
			v, err := s.products.GetVariantForUpdate(ctx, it.VariantID) // ล็อกแถวกัน oversell
			if err != nil {
				return err
			}
			if err := v.Reserve(it.Quantity); err != nil {
				return err
			}
			if err := s.products.SaveVariantStock(ctx, v); err != nil {
				return err
			}
			// snapshot ราคา "ปัจจุบัน" จาก variant ที่ล็อกไว้ ไม่ใช่ราคาที่ค้างในตะกร้า
			line := v.Price.Mul(decimal.NewFromInt(int64(it.Quantity)))
			items = append(items, domain.OrderItem{
				VariantID:   it.VariantID,
				ProductName: it.ProductName,
				VariantName: variantLabel(it.VariantName, it.Color),
				UnitPrice:   v.Price,
				Quantity:    it.Quantity,
				LineTotal:   line,
			})
			subtotal = subtotal.Add(line)
		}

		order = &domain.Order{
			UserID:         userID,
			Subtotal:       subtotal,
			ShippingFee:    flatShippingFee,
			DiscountAmount: decimal.Zero,
			TotalAmount:    subtotal.Add(flatShippingFee),
			Status:         domain.OrderPending,
			PaymentStatus:  domain.PaymentUnpaid,
			PaymentMethod:  cmd.PaymentMethod,
			Items:          items,
			Shipping:       domain.SnapshotAddress(addr),
			CustomerNote:   cmd.CustomerNote,
		}
		if err := s.orders.Create(ctx, order); err != nil {
			return err
		}
		return s.carts.MarkConverted(ctx, cart.ID)
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) List(ctx context.Context, userID uint, q input.OrderListQuery) ([]domain.Order, int, error) {
	limit := q.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}
	return s.orders.ListByUser(ctx, userID, output.OrderListFilter{
		Status: parseOrderStatus(q.Status),
		Limit:  limit,
		Offset: offset,
	})
}

func (s *OrderService) Get(ctx context.Context, userID uint, orderNumber string) (*domain.Order, error) {
	return s.orders.FindByNumber(ctx, userID, orderNumber)
}

// Cancel: ยกเลิก + คืนสต็อกทุกรายการใน tx เดียว (ล็อกแถว variant ก่อนบวกคืน)
func (s *OrderService) Cancel(ctx context.Context, userID uint, orderNumber, reason string) (*domain.Order, error) {
	var order *domain.Order
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		o, err := s.orders.FindByNumber(ctx, userID, orderNumber)
		if err != nil {
			return err
		}
		if !o.CanCancel() {
			return domain.ErrOrderNotCancellable
		}

		for _, it := range o.Items {
			if it.VariantID == 0 {
				continue // variant ถูกลบไปแล้ว (FK SET NULL) — ไม่มีสต็อกให้คืน
			}
			v, err := s.products.GetVariantForUpdate(ctx, it.VariantID)
			if err != nil {
				continue // variant หายไป — ข้ามการคืนสต็อกตัวนั้น ไม่ให้การยกเลิกล้ม
			}
			v.Restock(it.Quantity)
			if err := s.products.SaveVariantStock(ctx, v); err != nil {
				return err
			}
		}

		if err := s.orders.UpdateStatus(ctx, o.ID, domain.OrderCancelled, o.PaymentStatus, reason, 0); err != nil {
			return err
		}
		o.Status = domain.OrderCancelled
		order = o
		return nil
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

// SubmitPayment: แนบสลิป → payment สถานะ pending_review + order payment_status เปลี่ยนตาม
func (s *OrderService) SubmitPayment(ctx context.Context, userID uint, orderNumber string, cmd input.SubmitPaymentCommand) (*domain.Order, error) {
	if !domain.ValidPaymentMethod(cmd.Method) {
		return nil, domain.ErrInvalidInput
	}

	o, err := s.orders.FindByNumber(ctx, userID, orderNumber)
	if err != nil {
		return nil, err
	}
	if !o.CanSubmitPayment() {
		return nil, domain.ErrPaymentNotAllowed
	}
	// ยอดโอนต้องเท่ายอดออเดอร์เป๊ะ — โอนขาด/เกินให้ติดต่อร้านแทนที่จะรับเข้าระบบ
	if !cmd.Amount.Equal(o.TotalAmount) {
		return nil, domain.ErrAmountMismatch
	}

	// เซฟสลิปก่อนเปิด tx (I/O ช้าไม่ควรค้างใน tx) — ถ้า tx พังค่อยลบทิ้ง
	slipURL, err := s.files.Save(ctx,
		fmt.Sprintf("slips/%s-%d%s", o.OrderNumber, time.Now().UnixMilli(), cmd.SlipExt),
		cmd.SlipContent)
	if err != nil {
		return nil, err
	}

	payment := &domain.Payment{
		OrderID:        o.ID,
		Method:         cmd.Method,
		Amount:         cmd.Amount,
		Status:         domain.PaymentPendingReview,
		SlipURL:        slipURL,
		TransactionRef: cmd.TransactionRef,
	}
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.orders.CreatePayment(ctx, payment); err != nil {
			return err
		}
		return s.orders.UpdateStatus(ctx, o.ID, o.Status, domain.PaymentPendingReview, "ลูกค้าแจ้งชำระเงิน", 0)
	})
	if err != nil {
		_ = s.files.Remove(ctx, slipURL) // ไม่ทิ้งไฟล์กำพร้า
		return nil, err
	}

	o.PaymentStatus = domain.PaymentPendingReview
	o.PaymentMethod = cmd.Method
	o.Payment = payment
	return o, nil
}

// variantLabel รวมไซซ์+สีเป็นข้อความเดียวสำหรับ snapshot ใน order_items
func variantLabel(name, color string) string {
	if color == "" {
		return name
	}
	return name + " / " + color
}

// parseOrderStatus แปลง string → OrderStatus เฉพาะค่าที่รู้จัก (ค่าแปลก ๆ = ไม่กรอง)
func parseOrderStatus(s string) domain.OrderStatus {
	switch st := domain.OrderStatus(s); st {
	case domain.OrderPending, domain.OrderPaid, domain.OrderPreparing, domain.OrderShipped,
		domain.OrderDelivered, domain.OrderCompleted, domain.OrderCancelled, domain.OrderRefunded:
		return st
	}
	return ""
}
