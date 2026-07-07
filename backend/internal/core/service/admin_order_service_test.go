package service

import (
	"context"
	"errors"
	"testing"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

const adminID = uint(77)

// adminHarness ต่อยอด orderHarness: ลูกค้าสั่งซื้อ + แจ้งโอนไว้ก่อน แล้วให้แอดมินจัดการต่อ
type adminHarness struct {
	*orderHarness
	admin *AdminOrderService
}

func newAdminHarness(stock int) *adminHarness {
	h := newOrderHarness(stock)
	return &adminHarness{
		orderHarness: h,
		admin:        NewAdminOrderService(h.orders, h.products, fakeTx{}),
	}
}

// placeWithPayment สร้างออเดอร์ + แจ้งโอนจนสถานะ pending_review
func (h *adminHarness) placeWithPayment(t *testing.T, qty int) *domain.Order {
	t.Helper()
	addr := h.seedCheckout(qty)
	o, err := h.svc.Place(context.Background(), orderUID, placeCmd(addr.ID))
	if err != nil {
		t.Fatalf("place: %v", err)
	}
	o, err = h.svc.SubmitPayment(context.Background(), orderUID, o.OrderNumber, input.SubmitPaymentCommand{
		Method: domain.MethodPromptPay, Amount: o.TotalAmount, SlipContent: []byte("slip"), SlipExt: ".jpg",
	})
	if err != nil {
		t.Fatalf("submit payment: %v", err)
	}
	return o
}

func TestAdminVerifyPayment_ApproveMarksOrderPaid(t *testing.T) {
	h := newAdminHarness(5)
	o := h.placeWithPayment(t, 1)

	got, err := h.admin.VerifyPayment(context.Background(), adminID, o.Payment.ID, domain.PaymentPaid)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got.Status != domain.OrderPaid || got.PaymentStatus != domain.PaymentPaid {
		t.Fatalf("order should be paid/paid, got %s/%s", got.Status, got.PaymentStatus)
	}
	if h.orders.payments[0].Status != domain.PaymentPaid {
		t.Fatalf("payment status = %s", h.orders.payments[0].Status)
	}
	// ต้องบันทึกว่าแอดมินคนไหนเป็นคนตรวจ/แก้
	if h.orders.lastVerifiedBy != adminID || h.orders.lastChangedBy != adminID {
		t.Fatalf("admin id should be recorded, got verify=%d change=%d", h.orders.lastVerifiedBy, h.orders.lastChangedBy)
	}
}

func TestAdminVerifyPayment_RejectKeepsOrderPending(t *testing.T) {
	h := newAdminHarness(5)
	o := h.placeWithPayment(t, 1)

	got, err := h.admin.VerifyPayment(context.Background(), adminID, o.Payment.ID, domain.PaymentFailed)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	// ปฏิเสธสลิป: ออเดอร์ยัง pending ให้ลูกค้าแจ้งโอนใหม่ได้ (CanSubmitPayment เปิดรับ failed)
	if got.Status != domain.OrderPending || got.PaymentStatus != domain.PaymentFailed {
		t.Fatalf("want pending/failed, got %s/%s", got.Status, got.PaymentStatus)
	}
	if !got.CanSubmitPayment() {
		t.Fatal("customer should be able to resubmit payment after rejection")
	}
}

func TestAdminVerifyPayment_OnlyPendingReview(t *testing.T) {
	h := newAdminHarness(5)
	o := h.placeWithPayment(t, 1)
	if _, err := h.admin.VerifyPayment(context.Background(), adminID, o.Payment.ID, domain.PaymentPaid); err != nil {
		t.Fatalf("first verify: %v", err)
	}
	// ตัดสินซ้ำต้องโดนปฏิเสธ — กันเขียนทับผลเดิม
	if _, err := h.admin.VerifyPayment(context.Background(), adminID, o.Payment.ID, domain.PaymentFailed); !errors.Is(err, domain.ErrPaymentNotAllowed) {
		t.Fatalf("want ErrPaymentNotAllowed, got %v", err)
	}
}

func TestAdminVerifyPayment_InvalidStatus(t *testing.T) {
	h := newAdminHarness(5)
	if _, err := h.admin.VerifyPayment(context.Background(), adminID, 1, domain.PaymentPendingReview); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestAdminUpdateStatus_CancelRestocks(t *testing.T) {
	h := newAdminHarness(5)
	addr := h.seedCheckout(2)
	o, _ := h.svc.Place(context.Background(), orderUID, placeCmd(addr.ID))

	got, err := h.admin.UpdateStatus(context.Background(), adminID, o.OrderNumber, domain.OrderCancelled, "ของหมดหน้าร้าน")
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if got.Status != domain.OrderCancelled {
		t.Fatalf("status = %s", got.Status)
	}
	if stock := h.products.variants[100].StockQuantity; stock != 5 {
		t.Fatalf("stock should be restored to 5, got %d", stock)
	}
}

func TestAdminUpdateStatus_RejectsInvalidStatus(t *testing.T) {
	h := newAdminHarness(5)
	addr := h.seedCheckout(1)
	o, _ := h.svc.Place(context.Background(), orderUID, placeCmd(addr.ID))

	// pending เป็นสถานะตั้งต้นจากระบบ แอดมินย้อนกลับไปเองไม่ได้
	if _, err := h.admin.UpdateStatus(context.Background(), adminID, o.OrderNumber, domain.OrderPending, ""); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestAdminUpdateStatus_CancelAfterShippedRejected(t *testing.T) {
	h := newAdminHarness(5)
	addr := h.seedCheckout(1)
	o, _ := h.svc.Place(context.Background(), orderUID, placeCmd(addr.ID))
	h.orders.byNumber[o.OrderNumber].Status = domain.OrderShipped

	if _, err := h.admin.UpdateStatus(context.Background(), adminID, o.OrderNumber, domain.OrderCancelled, ""); !errors.Is(err, domain.ErrOrderNotCancellable) {
		t.Fatalf("want ErrOrderNotCancellable, got %v", err)
	}
}
