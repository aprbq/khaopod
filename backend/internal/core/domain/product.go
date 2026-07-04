package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// แคตตาล็อกสินค้า — struct ล้วน ไม่มี tag ของ gorm (persistence model แยกอยู่ใน postgres adapter)
// เงินใช้ decimal.Decimal เท่านั้น ห้าม float

// Category = หมวดหมู่สินค้า
type Category struct {
	ID       uint
	ParentID *uint
	Name     string
	Slug     string
}

// ProductImage = รูปสินค้า (มีได้หลายรูป, รูปปก is_primary ได้ 1 รูป)
type ProductImage struct {
	ID        uint
	ProductID uint
	URL       string
	AltText   string
	IsPrimary bool
	SortOrder int
}

// ProductVariant = ตัวเลือกสินค้า (ไซซ์/สี) ที่ถือราคาและสต็อกจริง
type ProductVariant struct {
	ID            uint
	ProductID     uint
	SKU           string
	Name          string // variant_name เช่น "ไซซ์ M / สีดำ"
	Price         decimal.Decimal
	StockQuantity int
	IsActive      bool
}

// Product = สินค้าหลัก — ราคา/สต็อกที่ขายจริงอยู่ที่ Variants
type Product struct {
	ID          uint
	CategoryID  *uint
	Category    *Category // อาจ nil ถ้าไม่ถูกจัดหมวด
	Name        string
	Slug        string
	Description string
	BasePrice   decimal.Decimal
	IsActive    bool
	IsFeatured  bool
	Variants    []ProductVariant
	Images      []ProductImage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// InStock = มี variant ที่ยังขายได้และมีสต็อกเหลืออย่างน้อยหนึ่งตัว
func (p *Product) InStock() bool {
	for _, v := range p.Variants {
		if v.IsActive && v.StockQuantity > 0 {
			return true
		}
	}
	return false
}

// PriceRange = ช่วงราคาต่ำสุด–สูงสุดจาก variant ที่ยังขายได้
// ถ้าไม่มี variant ที่ active จะ fallback เป็น base_price ทั้งคู่ (กันคืนค่าว่าง)
func (p *Product) PriceRange() (min, max decimal.Decimal) {
	first := true
	for _, v := range p.Variants {
		if !v.IsActive {
			continue
		}
		if first {
			min, max = v.Price, v.Price
			first = false
			continue
		}
		if v.Price.LessThan(min) {
			min = v.Price
		}
		if v.Price.GreaterThan(max) {
			max = v.Price
		}
	}
	if first {
		return p.BasePrice, p.BasePrice
	}
	return min, max
}

// PrimaryImage = URL ของรูปปก (is_primary) ถ้าไม่มีจะใช้รูปแรกตาม sort_order
// คืน "" ถ้าไม่มีรูปเลย (adapter จะ map เป็น null)
func (p *Product) PrimaryImage() string {
	if len(p.Images) == 0 {
		return ""
	}
	best := p.Images[0]
	for _, img := range p.Images {
		if img.IsPrimary {
			return img.URL
		}
		if img.SortOrder < best.SortOrder {
			best = img
		}
	}
	return best.URL
}
