package service

import (
	"context"
	"strings"

	"github.com/khaopod/backend/internal/core/domain"
	"github.com/khaopod/backend/internal/core/port/input"
	"github.com/khaopod/backend/internal/core/port/output"
)

// sortWhitelist = map จากค่า sort ที่ client ส่งได้ → ชื่อคอลัมน์จริงใน DB
// จำกัดไว้เพื่อกัน SQL injection ผ่านพารามิเตอร์ sort (ห้ามส่งชื่อคอลัมน์ดิบเข้า query)
var sortWhitelist = map[string]string{
	"created_at": "created_at",
	"base_price": "base_price",
	"name":       "name",
}

// ProductService = use case ของแคตตาล็อกสินค้า (implements input.ProductUseCase)
type ProductService struct {
	products output.ProductRepository
}

var _ input.ProductUseCase = (*ProductService)(nil)

func NewProductService(products output.ProductRepository) *ProductService {
	return &ProductService{products: products}
}

func (s *ProductService) List(ctx context.Context, q input.ProductQuery) ([]domain.Product, int, error) {
	page, perPage := q.NormalizedPaging()
	col, desc := parseSort(q.Sort)

	return s.products.List(ctx, output.ProductFilter{
		CategorySlug: strings.TrimSpace(q.CategorySlug),
		Search:       strings.TrimSpace(q.Search),
		SortColumn:   col,
		SortDesc:     desc,
		Limit:        perPage,
		Offset:       (page - 1) * perPage,
	})
}

func (s *ProductService) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, domain.ErrNotFound
	}
	return s.products.FindBySlug(ctx, slug)
}

func (s *ProductService) Categories(ctx context.Context) ([]domain.Category, error) {
	return s.products.ListCategories(ctx)
}

// parseSort แปลงค่า sort ("-created_at") เป็นคอลัมน์ + ทิศทาง
// prefix "-" = จากมากไปน้อย; ค่าที่ไม่อยู่ใน whitelist → default (created_at desc, ใหม่สุดก่อน)
func parseSort(sort string) (column string, desc bool) {
	sort = strings.TrimSpace(sort)
	desc = strings.HasPrefix(sort, "-")
	key := strings.TrimPrefix(sort, "-")
	col, ok := sortWhitelist[key]
	if !ok {
		return "created_at", true
	}
	return col, desc
}
