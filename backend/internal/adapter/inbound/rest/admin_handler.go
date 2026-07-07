package rest

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/khaopod/backend/internal/adapter/inbound/rest/response"
	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
)

// AdminHandler แปลง HTTP ↔ use case ฝั่งหลังบ้าน (§11 + จัดการแคตตาล็อก §5.3)
// ทุก route ถูกครอบด้วย RequireAuth + RequireAdmin ที่ router
type AdminHandler struct {
	admin   input.AdminOrderUseCase
	catalog input.AdminCatalogUseCase
	users   input.AdminUserUseCase
}

func NewAdminHandler(admin input.AdminOrderUseCase, catalog input.AdminCatalogUseCase, users input.AdminUserUseCase) *AdminHandler {
	return &AdminHandler{admin: admin, catalog: catalog, users: users}
}

// GET /admin/dashboard/summary (🛡️)
func (h *AdminHandler) Summary(c *gin.Context) {
	s, err := h.admin.Summary(c.Request.Context())
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, adminSummaryResponse{
		OrdersTotal:           s.OrdersTotal,
		OrdersPending:         s.OrdersPending,
		PaymentsPendingReview: s.PaymentsPendingReview,
		RevenuePaid:           s.RevenuePaid,
	})
}

// GET /admin/orders (🛡️) — ออเดอร์ของทุกคน
func (h *AdminHandler) ListOrders(c *gin.Context) {
	page := atoiDefault(c.Query("page"), 1)
	if page < 1 {
		page = 1
	}
	perPage := atoiDefault(c.Query("per_page"), 20)
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	items, total, err := h.admin.ListOrders(c.Request.Context(), input.OrderListQuery{
		Status: c.Query("status"),
		Limit:  perPage,
		Offset: (page - 1) * perPage,
	})
	if err != nil {
		mapError(c, err)
		return
	}

	out := make([]adminOrderResponse, 0, len(items))
	for i := range items {
		out = append(out, toAdminOrderResponse(&items[i]))
	}
	response.List(c, out, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages(total, perPage),
	})
}

// GET /admin/orders/:orderNumber (🛡️)
func (h *AdminHandler) GetOrder(c *gin.Context) {
	o, err := h.admin.GetOrder(c.Request.Context(), c.Param("orderNumber"))
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toAdminOrderResponse(o))
}

// PATCH /admin/orders/:orderNumber/status (🛡️)
func (h *AdminHandler) UpdateStatus(c *gin.Context) {
	var in updateOrderStatusRequest
	if !bindAndValidate(c, &in) {
		return
	}
	o, err := h.admin.UpdateStatus(c.Request.Context(), c.GetUint(ctxUserID),
		c.Param("orderNumber"), domain.OrderStatus(in.Status), in.Note)
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toAdminOrderResponse(o))
}

// PATCH /admin/payments/:id/verify (🛡️) — body: {"status":"paid"|"failed"}
func (h *AdminHandler) VerifyPayment(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var in verifyPaymentRequest
	if !bindAndValidate(c, &in) {
		return
	}
	o, err := h.admin.VerifyPayment(c.Request.Context(), c.GetUint(ctxUserID),
		id, domain.PaymentStatus(in.Status))
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toAdminOrderResponse(o))
}

// ---- Users ----

// GET /admin/users (🛡️)
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page := atoiDefault(c.Query("page"), 1)
	if page < 1 {
		page = 1
	}
	perPage := atoiDefault(c.Query("per_page"), 20)
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	items, total, err := h.users.ListUsers(c.Request.Context(), perPage, (page-1)*perPage)
	if err != nil {
		mapError(c, err)
		return
	}
	out := make([]adminUserResponse, 0, len(items))
	for i := range items {
		out = append(out, toAdminUserResponse(&items[i]))
	}
	response.List(c, out, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages(total, perPage),
	})
}

// ---- Catalog ----

// GET /admin/products (🛡️) — รวมสินค้าที่ปิดขาย
func (h *AdminHandler) ListProducts(c *gin.Context) {
	q := input.ProductQuery{
		CategorySlug: c.Query("category"),
		Search:       c.Query("search"),
		Sort:         c.Query("sort"),
		Page:         atoiDefault(c.Query("page"), 0),
		PerPage:      atoiDefault(c.Query("per_page"), 0),
	}
	items, total, err := h.catalog.ListProducts(c.Request.Context(), q)
	if err != nil {
		mapError(c, err)
		return
	}
	page, perPage := q.NormalizedPaging()
	out := make([]adminProductResponse, 0, len(items))
	for i := range items {
		out = append(out, toAdminProductResponse(&items[i]))
	}
	response.List(c, out, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages(total, perPage),
	})
}

// GET /admin/products/:id (🛡️)
func (h *AdminHandler) GetProduct(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	p, err := h.catalog.GetProduct(c.Request.Context(), id)
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toAdminProductResponse(p))
}

// POST /admin/products (🛡️)
func (h *AdminHandler) CreateProduct(c *gin.Context) {
	var in productRequest
	if !bindAndValidate(c, &in) {
		return
	}
	p, err := h.catalog.CreateProduct(c.Request.Context(), in.toCommand())
	if err != nil {
		mapError(c, err)
		return
	}
	response.Created(c, toAdminProductResponse(p))
}

// PATCH /admin/products/:id (🛡️)
func (h *AdminHandler) UpdateProduct(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var in productRequest
	if !bindAndValidate(c, &in) {
		return
	}
	p, err := h.catalog.UpdateProduct(c.Request.Context(), id, in.toCommand())
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toAdminProductResponse(p))
}

// DELETE /admin/products/:id (🛡️)
func (h *AdminHandler) DeleteProduct(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.catalog.DeleteProduct(c.Request.Context(), id); err != nil {
		mapError(c, err)
		return
	}
	response.NoContent(c)
}

// POST /admin/products/:id/variants (🛡️)
func (h *AdminHandler) CreateVariant(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var in variantRequest
	if !bindAndValidate(c, &in) {
		return
	}
	if err := h.catalog.CreateVariant(c.Request.Context(), id, in.toCommand()); err != nil {
		mapError(c, err)
		return
	}
	response.Created(c, gin.H{"ok": true})
}

// PATCH /admin/variants/:id (🛡️)
func (h *AdminHandler) UpdateVariant(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var in variantRequest
	if !bindAndValidate(c, &in) {
		return
	}
	if err := h.catalog.UpdateVariant(c.Request.Context(), id, in.toCommand()); err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, gin.H{"ok": true})
}

// DELETE /admin/variants/:id (🛡️)
func (h *AdminHandler) DeleteVariant(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.catalog.DeleteVariant(c.Request.Context(), id); err != nil {
		mapError(c, err)
		return
	}
	response.NoContent(c)
}

// POST /admin/products/:id/images (🛡️) — multipart field "image" (JPG/PNG/WebP ≤ 5MB)
func (h *AdminHandler) AddProductImage(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	content, ext, ok := readImageUpload(c, "image", maxSlipBytes)
	if !ok {
		return
	}
	if err := h.catalog.AddImage(c.Request.Context(), id, content, ext); err != nil {
		mapError(c, err)
		return
	}
	response.Created(c, gin.H{"ok": true})
}

// DELETE /admin/images/:id (🛡️)
func (h *AdminHandler) DeleteProductImage(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.catalog.DeleteImage(c.Request.Context(), id); err != nil {
		mapError(c, err)
		return
	}
	response.NoContent(c)
}

// POST /admin/products/:id/images/:imageId/primary (🛡️)
func (h *AdminHandler) SetPrimaryImage(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	imgID, err := strconv.ParseUint(c.Param("imageId"), 10, 64)
	if err != nil || imgID == 0 {
		response.Error(c, 400, response.CodeBadRequest, "รหัสรูปไม่ถูกต้อง")
		return
	}
	if err := h.catalog.SetPrimaryImage(c.Request.Context(), id, uint(imgID)); err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, gin.H{"ok": true})
}
