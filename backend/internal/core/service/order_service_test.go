package service

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

// orderHarness ประกอบ OrderService พร้อม fake ทุก dependency
type orderHarness struct {
	orders   *fakeOrderRepo
	carts    *fakeCartRepo
	products *fakeProductRepo
	addrs    *fakeAddressRepo
	files    *fakeFileStorage
	svc      *OrderService
}

func newOrderHarness(stock int) *orderHarness {
	variants := map[uint]domain.ProductVariant{
		100: {ID: 100, Name: "ไซซ์ M", Color: "ดำ", Price: decimal.NewFromInt(299), StockQuantity: stock, IsActive: true},
	}
	h := &orderHarness{
		orders:   newFakeOrderRepo(),
		carts:    newFakeCartRepo(variants),
		products: &fakeProductRepo{variants: variants},
		addrs:    newFakeAddressRepo(),
		files:    &fakeFileStorage{},
	}
	h.svc = NewOrderService(h.orders, h.carts, h.products, h.addrs, h.files, fakeTx{})
	return h
}

const orderUID = uint(1)

func (h *orderHarness) seedCheckout(qty int) *domain.Address {
	_ = h.carts.UpsertItem(context.Background(), orderUID, 100, qty)
	return h.addrs.seed(&domain.Address{
		UserID: orderUID, RecipientName: "สมชาย", Phone: "0812345678",
		AddressLine: "99/1", Subdistrict: "ในเมือง", District: "เมือง",
		Province: "ขอนแก่น", PostalCode: "40000", Country: "TH", IsDefault: true,
	})
}

func placeCmd(addrID uint) input.PlaceOrderCommand {
	return input.PlaceOrderCommand{AddressID: addrID, PaymentMethod: domain.MethodPromptPay}
}

func TestPlaceOrder_Success(t *testing.T) {
	h := newOrderHarness(5)
	addr := h.seedCheckout(2)

	o, err := h.svc.Place(context.Background(), orderUID, placeCmd(addr.ID))
	if err != nil {
		t.Fatalf("place: %v", err)
	}
	// ยอด: 299×2 + ค่าส่ง 40
	if !o.Subtotal.Equal(decimal.NewFromInt(598)) || !o.TotalAmount.Equal(decimal.NewFromInt(638)) {
		t.Fatalf("wrong totals: subtotal=%s total=%s", o.Subtotal, o.TotalAmount)
	}
	// ตัดสต็อกจริง
	if got := h.products.variants[100].StockQuantity; got != 3 {
		t.Fatalf("stock should be deducted to 3, got %d", got)
	}
	// snapshot รวมไซซ์+สี และที่อยู่
	if o.Items[0].VariantName != "ไซซ์ M / ดำ" {
		t.Fatalf("variant label = %q", o.Items[0].VariantName)
	}
	if o.Shipping.Recipient != "สมชาย" || o.Shipping.PostalCode != "40000" {
		t.Fatalf("shipping snapshot wrong: %+v", o.Shipping)
	}
	// ตะกร้าถูกปิด
	if h.carts.convertedCartID == 0 {
		t.Fatal("cart should be marked converted")
	}
	if o.Status != domain.OrderPending || o.PaymentStatus != domain.PaymentUnpaid {
		t.Fatalf("wrong initial status: %s/%s", o.Status, o.PaymentStatus)
	}
}

func TestPlaceOrder_OutOfStock(t *testing.T) {
	h := newOrderHarness(1)
	addr := h.seedCheckout(2) // สั่ง 2 แต่สต็อกมี 1

	_, err := h.svc.Place(context.Background(), orderUID, placeCmd(addr.ID))
	if !errors.Is(err, domain.ErrOutOfStock) {
		t.Fatalf("want ErrOutOfStock, got %v", err)
	}
}

func TestPlaceOrder_EmptyCart(t *testing.T) {
	h := newOrderHarness(5)
	addr := h.addrs.seed(&domain.Address{UserID: orderUID, RecipientName: "สมชาย"})

	_, err := h.svc.Place(context.Background(), orderUID, placeCmd(addr.ID))
	if !errors.Is(err, domain.ErrCartEmpty) {
		t.Fatalf("want ErrCartEmpty, got %v", err)
	}
}

func TestPlaceOrder_AddressOfOtherUser(t *testing.T) {
	h := newOrderHarness(5)
	h.seedCheckout(1)
	other := h.addrs.seed(&domain.Address{UserID: 99, RecipientName: "คนอื่น"})

	// ใช้ address ของคนอื่นต้องไม่ได้ (IDOR)
	_, err := h.svc.Place(context.Background(), orderUID, placeCmd(other.ID))
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestPlaceOrder_InvalidMethod(t *testing.T) {
	h := newOrderHarness(5)
	addr := h.seedCheckout(1)

	cmd := input.PlaceOrderCommand{AddressID: addr.ID, PaymentMethod: "bitcoin"}
	if _, err := h.svc.Place(context.Background(), orderUID, cmd); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestCancelOrder_RestocksItems(t *testing.T) {
	h := newOrderHarness(5)
	addr := h.seedCheckout(2)
	o, err := h.svc.Place(context.Background(), orderUID, placeCmd(addr.ID))
	if err != nil {
		t.Fatalf("place: %v", err)
	}

	got, err := h.svc.Cancel(context.Background(), orderUID, o.OrderNumber, "สั่งผิด")
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if got.Status != domain.OrderCancelled {
		t.Fatalf("status = %s", got.Status)
	}
	// สต็อกกลับมาเท่าเดิม (5-2+2)
	if stock := h.products.variants[100].StockQuantity; stock != 5 {
		t.Fatalf("stock should be restored to 5, got %d", stock)
	}
}

func TestCancelOrder_NotCancellableAfterShipped(t *testing.T) {
	h := newOrderHarness(5)
	addr := h.seedCheckout(1)
	o, _ := h.svc.Place(context.Background(), orderUID, placeCmd(addr.ID))
	h.orders.byNumber[o.OrderNumber].Status = domain.OrderShipped

	_, err := h.svc.Cancel(context.Background(), orderUID, o.OrderNumber, "")
	if !errors.Is(err, domain.ErrOrderNotCancellable) {
		t.Fatalf("want ErrOrderNotCancellable, got %v", err)
	}
}

func TestSubmitPayment_Success(t *testing.T) {
	h := newOrderHarness(5)
	addr := h.seedCheckout(2)
	o, _ := h.svc.Place(context.Background(), orderUID, placeCmd(addr.ID))

	got, err := h.svc.SubmitPayment(context.Background(), orderUID, o.OrderNumber, input.SubmitPaymentCommand{
		Method:      domain.MethodBankTransfer,
		Amount:      o.TotalAmount,
		SlipContent: []byte("slip"),
		SlipExt:     ".jpg",
	})
	if err != nil {
		t.Fatalf("submit payment: %v", err)
	}
	if got.PaymentStatus != domain.PaymentPendingReview {
		t.Fatalf("payment status = %s", got.PaymentStatus)
	}
	if got.Payment == nil || got.Payment.SlipURL == "" {
		t.Fatal("payment with slip url expected")
	}
	if len(h.files.saved) != 1 {
		t.Fatalf("slip file should be saved once, got %d", len(h.files.saved))
	}
}

func TestSubmitPayment_AmountMismatch(t *testing.T) {
	h := newOrderHarness(5)
	addr := h.seedCheckout(1)
	o, _ := h.svc.Place(context.Background(), orderUID, placeCmd(addr.ID))

	_, err := h.svc.SubmitPayment(context.Background(), orderUID, o.OrderNumber, input.SubmitPaymentCommand{
		Method: domain.MethodPromptPay, Amount: decimal.NewFromInt(1), SlipContent: []byte("x"), SlipExt: ".jpg",
	})
	if !errors.Is(err, domain.ErrAmountMismatch) {
		t.Fatalf("want ErrAmountMismatch, got %v", err)
	}
}

func TestSubmitPayment_TwiceRejected(t *testing.T) {
	h := newOrderHarness(5)
	addr := h.seedCheckout(1)
	o, _ := h.svc.Place(context.Background(), orderUID, placeCmd(addr.ID))

	cmd := input.SubmitPaymentCommand{
		Method: domain.MethodPromptPay, Amount: o.TotalAmount, SlipContent: []byte("x"), SlipExt: ".jpg",
	}
	if _, err := h.svc.SubmitPayment(context.Background(), orderUID, o.OrderNumber, cmd); err != nil {
		t.Fatalf("first submit: %v", err)
	}
	// แจ้งซ้ำระหว่างรอตรวจต้องโดนปฏิเสธ
	if _, err := h.svc.SubmitPayment(context.Background(), orderUID, o.OrderNumber, cmd); !errors.Is(err, domain.ErrPaymentNotAllowed) {
		t.Fatalf("want ErrPaymentNotAllowed, got %v", err)
	}
}

func TestSubmitPayment_DBErrorCleansUpSlip(t *testing.T) {
	h := newOrderHarness(5)
	addr := h.seedCheckout(1)
	o, _ := h.svc.Place(context.Background(), orderUID, placeCmd(addr.ID))
	h.orders.paymentErr = errors.New("db down")

	_, err := h.svc.SubmitPayment(context.Background(), orderUID, o.OrderNumber, input.SubmitPaymentCommand{
		Method: domain.MethodPromptPay, Amount: o.TotalAmount, SlipContent: []byte("x"), SlipExt: ".jpg",
	})
	if err == nil {
		t.Fatal("want error when payment insert fails")
	}
	// สลิปที่เพิ่งเซฟต้องถูกลบทิ้ง
	if len(h.files.saved) != 0 {
		t.Fatalf("slip should be cleaned up, still have %v", h.files.saved)
	}
}

func TestGetOrder_OwnershipEnforced(t *testing.T) {
	h := newOrderHarness(5)
	addr := h.seedCheckout(1)
	o, _ := h.svc.Place(context.Background(), orderUID, placeCmd(addr.ID))

	// user อื่นดูออเดอร์เราไม่ได้
	if _, err := h.svc.Get(context.Background(), 99, o.OrderNumber); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
