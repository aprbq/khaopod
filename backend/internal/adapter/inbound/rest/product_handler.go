package rest

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/khaopod/backend/internal/adapter/inbound/rest/response"
	"github.com/khaopod/backend/internal/core/port/input"
)

// ProductHandler แปลง HTTP ↔ ProductUseCase (แคตตาล็อกสินค้าสาธารณะ 🔓)
type ProductHandler struct {
	products input.ProductUseCase
}

func NewProductHandler(products input.ProductUseCase) *ProductHandler {
	return &ProductHandler{products: products}
}

// GET /products (🔓) — filter/search/paginate
func (h *ProductHandler) List(c *gin.Context) {
	q := input.ProductQuery{
		CategorySlug: c.Query("category"),
		Search:       c.Query("search"),
		Sort:         c.Query("sort"),
		Page:         atoiDefault(c.Query("page"), 0),
		PerPage:      atoiDefault(c.Query("per_page"), 0),
	}

	items, total, err := h.products.List(c.Request.Context(), q)
	if err != nil {
		mapError(c, err)
		return
	}

	// meta ต้องสะท้อนค่าที่ใช้จริงหลัง clamp (per_page เกินเพดานจะถูกลด)
	page, perPage := q.NormalizedPaging()
	out := make([]productListItem, 0, len(items))
	for _, p := range items {
		out = append(out, toProductListItem(p))
	}
	response.List(c, out, response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages(total, perPage),
	})
}

// GET /products/{slug} (🔓) — รายละเอียด + variants + รูป
func (h *ProductHandler) GetBySlug(c *gin.Context) {
	p, err := h.products.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		mapError(c, err)
		return
	}
	response.OK(c, toProductDetailResponse(p))
}

// atoiDefault แปลง query string เป็น int, คืนค่า default ถ้าว่างหรือ parse ไม่ได้
func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func totalPages(total, perPage int) int {
	if perPage <= 0 || total <= 0 {
		return 0
	}
	return (total + perPage - 1) / perPage
}
