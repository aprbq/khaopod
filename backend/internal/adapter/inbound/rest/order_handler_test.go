package rest

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

// fakeOrderUC จำ command ที่ handler ส่งเข้ามา — เทสเฉพาะชั้น HTTP (validation + mapping)
type fakeOrderUC struct {
	gotPayment input.SubmitPaymentCommand
	placeErr   error
}

func (f *fakeOrderUC) Place(_ context.Context, _ uint, _ input.PlaceOrderCommand) (*domain.Order, error) {
	if f.placeErr != nil {
		return nil, f.placeErr
	}
	return &domain.Order{OrderNumber: "ORD-X", Status: domain.OrderPending}, nil
}
func (f *fakeOrderUC) List(context.Context, uint, input.OrderListQuery) ([]domain.Order, int, error) {
	return nil, 0, nil
}
func (f *fakeOrderUC) Get(context.Context, uint, string) (*domain.Order, error) {
	return &domain.Order{OrderNumber: "ORD-X"}, nil
}
func (f *fakeOrderUC) Cancel(context.Context, uint, string, string) (*domain.Order, error) {
	return &domain.Order{OrderNumber: "ORD-X", Status: domain.OrderCancelled}, nil
}
func (f *fakeOrderUC) SubmitPayment(_ context.Context, _ uint, _ string, cmd input.SubmitPaymentCommand) (*domain.Order, error) {
	f.gotPayment = cmd
	return &domain.Order{OrderNumber: "ORD-X", PaymentStatus: domain.PaymentPendingReview}, nil
}

func setupOrderHandler(uc input.OrderUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	h := NewOrderHandler(uc)
	e.POST("/orders", h.Place)
	e.POST("/orders/:orderNumber/payment", h.SubmitPayment)
	return e
}

func doPayment(e *gin.Engine, fields map[string]string, slipField string, slip []byte) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	if slip != nil {
		fw, _ := w.CreateFormFile(slipField, "slip.png")
		_, _ = fw.Write(slip)
	}
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/orders/ORD-X/payment", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestSubmitPayment_AcceptsValidSlip(t *testing.T) {
	uc := &fakeOrderUC{}
	rec := doPayment(setupOrderHandler(uc),
		map[string]string{"method": "bank_transfer", "amount": "638.00", "transaction_ref": "TX1"},
		"slip", pngBytes)
	if rec.Code != 201 {
		t.Fatalf("want 201, got %d (%s)", rec.Code, rec.Body.String())
	}
	if uc.gotPayment.Method != domain.MethodBankTransfer || uc.gotPayment.SlipExt != ".png" {
		t.Fatalf("command wrong: %+v", uc.gotPayment)
	}
	if !uc.gotPayment.Amount.Equal(decimal.RequireFromString("638.00")) {
		t.Fatalf("amount = %s", uc.gotPayment.Amount)
	}
}

func TestSubmitPayment_RejectsBadMethod(t *testing.T) {
	rec := doPayment(setupOrderHandler(&fakeOrderUC{}),
		map[string]string{"method": "bitcoin", "amount": "1"}, "slip", pngBytes)
	if rec.Code != 422 {
		t.Fatalf("want 422, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestSubmitPayment_RejectsMissingSlip(t *testing.T) {
	rec := doPayment(setupOrderHandler(&fakeOrderUC{}),
		map[string]string{"method": "promptpay", "amount": "100"}, "slip", nil)
	if rec.Code != 400 {
		t.Fatalf("want 400, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestSubmitPayment_RejectsNonImageSlip(t *testing.T) {
	rec := doPayment(setupOrderHandler(&fakeOrderUC{}),
		map[string]string{"method": "promptpay", "amount": "100"}, "slip", []byte("not an image at all"))
	if rec.Code != 422 {
		t.Fatalf("want 422, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestPlaceOrder_MapsCartEmpty(t *testing.T) {
	e := setupOrderHandler(&fakeOrderUC{placeErr: domain.ErrCartEmpty})
	rec := doJSON(e, "/orders", `{"address_id":1,"payment_method":"promptpay"}`)
	if rec.Code != 409 {
		t.Fatalf("want 409, got %d (%s)", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("CART_EMPTY")) {
		t.Fatalf("want CART_EMPTY code, got %s", rec.Body.String())
	}
}
