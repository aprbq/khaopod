package service

import (
	"context"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
	"github.com/khaopod/backend/internal/core/port/output"
)

// AdminOrderService = use case หลังบ้านแอดมิน (implements input.AdminOrderUseCase)
// การเช็ค role เกิดที่ middleware ฝั่ง inbound — service วางใจว่า caller เป็นแอดมินแล้ว
type AdminOrderService struct {
	orders   output.OrderRepository
	products output.ProductRepository
	tx       output.TxManager
}

var _ input.AdminOrderUseCase = (*AdminOrderService)(nil)

func NewAdminOrderService(orders output.OrderRepository, products output.ProductRepository, tx output.TxManager) *AdminOrderService {
	return &AdminOrderService{orders: orders, products: products, tx: tx}
}

func (s *AdminOrderService) Summary(ctx context.Context) (*domain.AdminSummary, error) {
	return s.orders.Summary(ctx)
}

func (s *AdminOrderService) ListOrders(ctx context.Context, q input.OrderListQuery) ([]domain.Order, int, error) {
	limit := q.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}
	return s.orders.ListAll(ctx, output.OrderListFilter{
		Status: parseOrderStatus(q.Status),
		Limit:  limit,
		Offset: offset,
	})
}

func (s *AdminOrderService) GetOrder(ctx context.Context, orderNumber string) (*domain.Order, error) {
	return s.orders.FindByNumberAny(ctx, orderNumber)
}

// UpdateStatus เปลี่ยนสถานะ + บันทึกผู้แก้ — เปลี่ยนเป็น cancelled จะคืนสต็อกใน tx เดียวกัน
func (s *AdminOrderService) UpdateStatus(ctx context.Context, adminID uint, orderNumber string, status domain.OrderStatus, note string) (*domain.Order, error) {
	if !domain.ValidAdminStatus(status) {
		return nil, domain.ErrInvalidInput
	}

	var order *domain.Order
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		o, err := s.orders.FindByNumberAny(ctx, orderNumber)
		if err != nil {
			return err
		}

		if status == domain.OrderCancelled {
			// ยกเลิกโดยแอดมินใช้กติกาเดียวกับลูกค้า: ก่อนจัดส่งเท่านั้น + คืนสต็อก
			if !o.CanCancel() {
				return domain.ErrOrderNotCancellable
			}
			for _, it := range o.Items {
				if it.VariantID == 0 {
					continue
				}
				v, err := s.products.GetVariantForUpdate(ctx, it.VariantID)
				if err != nil {
					continue // variant ถูกลบไปแล้ว — ข้ามการคืนสต็อกตัวนั้น
				}
				v.Restock(it.Quantity)
				if err := s.products.SaveVariantStock(ctx, v); err != nil {
					return err
				}
			}
		}

		if err := s.orders.UpdateStatus(ctx, o.ID, status, o.PaymentStatus, note, adminID); err != nil {
			return err
		}
		o.Status = status
		order = o
		return nil
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

// VerifyPayment ตัดสินสลิปที่รอตรวจ — paid: ออเดอร์ขยับเป็นจ่ายแล้ว, failed: ให้ลูกค้าแจ้งใหม่ได้
func (s *AdminOrderService) VerifyPayment(ctx context.Context, adminID uint, paymentID uint, status domain.PaymentStatus) (*domain.Order, error) {
	if status != domain.PaymentPaid && status != domain.PaymentFailed {
		return nil, domain.ErrInvalidInput
	}

	var order *domain.Order
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		p, err := s.orders.FindPaymentByID(ctx, paymentID)
		if err != nil {
			return err
		}
		// ตัดสินได้เฉพาะสลิปที่ยังรอตรวจ — กันเขียนทับผลเดิม
		if p.Status != domain.PaymentPendingReview {
			return domain.ErrPaymentNotAllowed
		}

		if err := s.orders.SavePaymentVerdict(ctx, p.ID, status, adminID); err != nil {
			return err
		}

		o, err := s.orders.FindByIDAny(ctx, p.OrderID)
		if err != nil {
			return err
		}
		orderStatus := o.Status
		note := "ปฏิเสธสลิป — รอลูกค้าแจ้งชำระเงินใหม่"
		if status == domain.PaymentPaid {
			note = "ยืนยันการชำระเงินแล้ว"
			if o.Status == domain.OrderPending {
				orderStatus = domain.OrderPaid
			}
		}
		if err := s.orders.UpdateStatus(ctx, o.ID, orderStatus, status, note, adminID); err != nil {
			return err
		}
		o.Status = orderStatus
		o.PaymentStatus = status
		order = o
		return nil
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}
