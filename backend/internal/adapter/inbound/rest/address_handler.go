package rest

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/khaopod/backend/internal/adapter/inbound/rest/response"
	"github.com/khaopod/backend/internal/core/port/input"
)

// AddressHandler แปลง HTTP ↔ AddressUseCase (ที่อยู่จัดส่ง §7)
type AddressHandler struct {
	addrs input.AddressUseCase
}

func NewAddressHandler(addrs input.AddressUseCase) *AddressHandler {
	return &AddressHandler{addrs: addrs}
}

// GET /addresses (🔒)
func (h *AddressHandler) List(c *gin.Context) {
	items, err := h.addrs.List(c.Request.Context(), c.GetUint(ctxUserID))
	if err != nil {
		mapError(c, err)
		return
	}
	out := make([]addressResponse, 0, len(items))
	for i := range items {
		out = append(out, toAddressResponse(&items[i]))
	}
	response.OK(c, out)
}

// POST /addresses (🔒)
func (h *AddressHandler) Create(c *gin.Context) {
	var in addressRequest
	if !bindAndValidate(c, &in) {
		return
	}
	a, err := h.addrs.Create(c.Request.Context(), c.GetUint(ctxUserID), in.toCommand())
	if err != nil {
		mapError(c, err)
		return
	}
	response.Created(c, toAddressResponse(a))
}

// GET /addresses/:id (🔒)
func (h *AddressHandler) Get(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	a, err := h.addrs.Get(c.Request.Context(), c.GetUint(ctxUserID), id)
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toAddressResponse(a))
}

// PATCH /addresses/:id (🔒)
func (h *AddressHandler) Update(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var in addressRequest
	if !bindAndValidate(c, &in) {
		return
	}
	a, err := h.addrs.Update(c.Request.Context(), c.GetUint(ctxUserID), id, in.toCommand())
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toAddressResponse(a))
}

// DELETE /addresses/:id (🔒)
func (h *AddressHandler) Delete(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.addrs.Delete(c.Request.Context(), c.GetUint(ctxUserID), id); err != nil {
		mapError(c, err)
		return
	}
	response.NoContent(c)
}

// POST /addresses/:id/default (🔒)
func (h *AddressHandler) SetDefault(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	a, err := h.addrs.SetDefault(c.Request.Context(), c.GetUint(ctxUserID), id)
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toAddressResponse(a))
}

// paramID อ่าน :id จาก path — ตอบ 400 เองถ้าไม่ใช่ตัวเลข (คืน ok=false)
func paramID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, 400, response.CodeBadRequest, "รหัสไม่ถูกต้อง")
		return 0, false
	}
	return uint(id), true
}
