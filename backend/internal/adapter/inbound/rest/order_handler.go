package rest

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/khaopod/backend/internal/adapter/inbound/rest/response"
	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

// เพดานขนาดไฟล์สลิป — ใหญ่กว่ารูปโปรไฟล์เพราะสลิปจากแอปธนาคารมักละเอียดกว่า
const maxSlipBytes = 5 << 20 // 5MB

// OrderHandler แปลง HTTP ↔ OrderUseCase (คำสั่งซื้อ §9 + ชำระเงิน §10)
type OrderHandler struct {
	orders input.OrderUseCase
}

func NewOrderHandler(orders input.OrderUseCase) *OrderHandler {
	return &OrderHandler{orders: orders}
}

// POST /orders (🔒) — checkout จากตะกร้าทั้งใบ
func (h *OrderHandler) Place(c *gin.Context) {
	var in placeOrderRequest
	if !bindAndValidate(c, &in) {
		return
	}
	o, err := h.orders.Place(c.Request.Context(), c.GetUint(ctxUserID), in.toCommand())
	if err != nil {
		mapError(c, err)
		return
	}
	response.Created(c, toOrderResponse(o))
}

// GET /orders (🔒) — รายการออเดอร์ของตัวเอง
func (h *OrderHandler) List(c *gin.Context) {
	page := atoiDefault(c.Query("page"), 1)
	if page < 1 {
		page = 1
	}
	perPage := atoiDefault(c.Query("per_page"), 20)
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	items, total, err := h.orders.List(c.Request.Context(), c.GetUint(ctxUserID), input.OrderListQuery{
		Status: c.Query("status"),
		Limit:  perPage,
		Offset: (page - 1) * perPage,
	})
	if err != nil {
		mapError(c, err)
		return
	}

	out := make([]orderResponse, 0, len(items))
	for i := range items {
		out = append(out, toOrderResponse(&items[i]))
	}
	response.List(c, out, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages(total, perPage),
	})
}

// GET /orders/:orderNumber (🔒)
func (h *OrderHandler) Get(c *gin.Context) {
	o, err := h.orders.Get(c.Request.Context(), c.GetUint(ctxUserID), c.Param("orderNumber"))
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toOrderResponse(o))
}

// POST /orders/:orderNumber/cancel (🔒)
func (h *OrderHandler) Cancel(c *gin.Context) {
	var in cancelOrderRequest
	_ = c.ShouldBindJSON(&in) // reason เป็น optional — body ว่างก็ยกเลิกได้

	o, err := h.orders.Cancel(c.Request.Context(), c.GetUint(ctxUserID), c.Param("orderNumber"), in.Reason)
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toOrderResponse(o))
}

// POST /orders/:orderNumber/payment (🔒) — multipart: method, amount, transaction_ref, slip
func (h *OrderHandler) SubmitPayment(c *gin.Context) {
	method := domain.PaymentMethod(c.PostForm("method"))
	if !domain.ValidPaymentMethod(method) {
		response.Error(c, 422, response.CodeValidation, "วิธีชำระเงินไม่ถูกต้อง (promptpay หรือ bank_transfer)")
		return
	}
	amount, err := decimal.NewFromString(c.PostForm("amount"))
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		response.Error(c, 422, response.CodeValidation, "ยอดโอนไม่ถูกต้อง")
		return
	}

	fh, err := c.FormFile("slip")
	if err != nil {
		response.Error(c, 400, response.CodeBadRequest, "กรุณาแนบรูปสลิปในฟิลด์ slip")
		return
	}
	if fh.Size > maxSlipBytes {
		response.Error(c, 422, response.CodeValidation, "ไฟล์สลิปต้องไม่เกิน 5MB")
		return
	}
	f, err := fh.Open()
	if err != nil {
		mapError(c, err)
		return
	}
	defer f.Close()
	content, err := io.ReadAll(io.LimitReader(f, maxSlipBytes+1))
	if err != nil {
		mapError(c, err)
		return
	}
	if len(content) > maxSlipBytes {
		response.Error(c, 422, response.CodeValidation, "ไฟล์สลิปต้องไม่เกิน 5MB")
		return
	}
	// ตรวจชนิดจาก magic bytes จริง — ไม่เชื่อนามสกุล/Content-Type ที่ client ส่งมา
	ext, ok := avatarExt(http.DetectContentType(content))
	if !ok {
		response.Error(c, 422, response.CodeValidation, "รองรับเฉพาะไฟล์ JPG, PNG หรือ WebP")
		return
	}

	o, err := h.orders.SubmitPayment(c.Request.Context(), c.GetUint(ctxUserID), c.Param("orderNumber"),
		input.SubmitPaymentCommand{
			Method:         method,
			Amount:         amount,
			TransactionRef: c.PostForm("transaction_ref"),
			SlipContent:    content,
			SlipExt:        ext,
		})
	if err != nil {
		mapError(c, err)
		return
	}
	response.Created(c, toOrderResponse(o))
}

// GET /orders/:orderNumber/payment (🔒) — สถานะการชำระเงินล่าสุดของออเดอร์
func (h *OrderHandler) GetPayment(c *gin.Context) {
	o, err := h.orders.Get(c.Request.Context(), c.GetUint(ctxUserID), c.Param("orderNumber"))
	if err != nil {
		mapError(c, err)
		return
	}
	if o.Payment == nil {
		response.Error(c, 404, response.CodeNotFound, "ยังไม่มีการแจ้งชำระเงินของออเดอร์นี้")
		return
	}
	response.OK(c, toPaymentResponse(o.Payment))
}
