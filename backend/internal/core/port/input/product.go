package input

import (
	"context"

	"github.com/khaopod/backend/internal/core/domain"
)

// ProductUseCase = ดูแคตตาล็อกสินค้าหน้าร้าน (driving port, สาธารณะ 🔓)
type ProductUseCase interface {
	// List คืนสินค้าที่ active ตาม filter พร้อมจำนวนทั้งหมด (ไว้คำนวณ pagination)
	List(ctx context.Context, q ProductQuery) ([]domain.Product, int, error)
	// GetBySlug คืนรายละเอียดสินค้า (พร้อม variants + รูป) — domain.ErrNotFound ถ้าไม่พบ/ไม่ active
	GetBySlug(ctx context.Context, slug string) (*domain.Product, error)
	// Categories คืนหมวดหมู่ที่ active ทั้งหมด (§4.1)
	Categories(ctx context.Context) ([]domain.Category, error)
}

// ProductQuery = พารามิเตอร์ค้นหา/แบ่งหน้า (ค่าดิบจาก query string)
type ProductQuery struct {
	CategorySlug string // filter ตาม slug ของหมวดหมู่ ("" = ทุกหมวด)
	Search       string // ค้นจากชื่อสินค้า
	Sort         string // เช่น "-created_at", "base_price", "name" (service คัดเฉพาะที่อนุญาต)
	Page         int
	PerPage      int
}

const (
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// NormalizedPaging คืนค่า page/per_page ที่ปรับให้อยู่ในกรอบที่ปลอดภัยแล้ว
// ใช้ทั้งใน service (สร้าง limit/offset) และ inbound adapter (สร้าง meta) เพื่อให้ตรงกันเสมอ
func (q ProductQuery) NormalizedPaging() (page, perPage int) {
	page = q.Page
	if page < 1 {
		page = 1
	}
	perPage = q.PerPage
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}
	return page, perPage
}
