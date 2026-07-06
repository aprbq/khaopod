package rest

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/khaopod/backend/internal/adapter/inbound/rest/response"
	"github.com/khaopod/backend/internal/core/port/input"
)

// CartHandler แปลง HTTP ↔ CartUseCase — ทุก endpoint ต้องล็อกอิน (🔒)
// user_id มาจาก auth middleware (c.GetUint(ctxUserID)) → เป็น scope กัน IDOR อัตโนมัติ
type CartHandler struct {
	carts input.CartUseCase
}

func NewCartHandler(carts input.CartUseCase) *CartHandler {
	return &CartHandler{carts: carts}
}

// GET /cart
func (h *CartHandler) Get(c *gin.Context) {
	cart, err := h.carts.Get(c.Request.Context(), c.GetUint(ctxUserID))
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toCartResponse(cart))
}

// POST /cart/items
func (h *CartHandler) AddItem(c *gin.Context) {
	var in addCartItemRequest
	if !bindAndValidate(c, &in) {
		return
	}
	cart, err := h.carts.AddItem(c.Request.Context(), c.GetUint(ctxUserID), in.toCommand())
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toCartResponse(cart))
}

// PATCH /cart/items/:itemId
func (h *CartHandler) UpdateItem(c *gin.Context) {
	itemID, ok := parseIDParam(c, "itemId")
	if !ok {
		return
	}
	var in updateCartItemRequest
	if !bindAndValidate(c, &in) {
		return
	}
	cart, err := h.carts.UpdateItem(c.Request.Context(), c.GetUint(ctxUserID), itemID, in.Quantity)
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toCartResponse(cart))
}

// DELETE /cart/items/:itemId
func (h *CartHandler) RemoveItem(c *gin.Context) {
	itemID, ok := parseIDParam(c, "itemId")
	if !ok {
		return
	}
	cart, err := h.carts.RemoveItem(c.Request.Context(), c.GetUint(ctxUserID), itemID)
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toCartResponse(cart))
}

// DELETE /cart
func (h *CartHandler) Clear(c *gin.Context) {
	if err := h.carts.Clear(c.Request.Context(), c.GetUint(ctxUserID)); err != nil {
		mapError(c, err)
		return
	}
	response.NoContent(c)
}

// parseIDParam อ่าน path param เป็น uint (>0) — ตอบ 400 ถ้าไม่ใช่ตัวเลขที่ถูกต้อง
func parseIDParam(c *gin.Context, name string) (uint, bool) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || v == 0 {
		response.Error(c, 400, response.CodeBadRequest, "รหัสไม่ถูกต้อง")
		return 0, false
	}
	return uint(v), true
}
