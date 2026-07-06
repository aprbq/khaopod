package domain

import "github.com/shopspring/decimal"

// ตะกร้าสินค้า — struct ล้วน ไม่มี tag ของ gorm
// ราคา/สต็อกใน CartItem เป็นค่า "ปัจจุบัน" (ไม่ใช่ snapshot) — cart โชว์ราคาสด ณ ตอนนี้
// (snapshot ราคาจะเกิดตอนสร้างออเดอร์ ไม่ใช่ที่นี่)

type CartItem struct {
	ID          uint
	VariantID   uint
	ProductName string
	VariantName string // ไซซ์
	Color       string // สี ("" = ไม่มีตัวเลือกสี)
	UnitPrice   decimal.Decimal
	Image       string // รูปปกของสินค้า ("" = ไม่มีรูป)
	Stock       int    // สต็อกปัจจุบันของ variant
	Quantity    int
}

// LineTotal = ราคารวมของรายการนี้ (unit_price × quantity)
func (i CartItem) LineTotal() decimal.Decimal {
	return i.UnitPrice.Mul(decimal.NewFromInt(int64(i.Quantity)))
}

// InStock = สต็อกยังพอกับจำนวนที่อยู่ในตะกร้า
func (i CartItem) InStock() bool { return i.Stock >= i.Quantity }

type Cart struct {
	ID     uint
	UserID uint
	Items  []CartItem
}

// Subtotal = ผลรวม line_total ทุกรายการ
func (c Cart) Subtotal() decimal.Decimal {
	sum := decimal.Zero
	for _, it := range c.Items {
		sum = sum.Add(it.LineTotal())
	}
	return sum
}

// ItemCount = จำนวนชิ้นรวมทั้งตะกร้า (ผลรวม quantity) — ใช้โชว์ badge
func (c Cart) ItemCount() int {
	n := 0
	for _, it := range c.Items {
		n += it.Quantity
	}
	return n
}

// ItemByVariant หา item ที่ตรงกับ variant (ไว้เช็คว่ามีอยู่แล้วก่อน merge)
func (c Cart) ItemByVariant(variantID uint) (CartItem, bool) {
	for _, it := range c.Items {
		if it.VariantID == variantID {
			return it, true
		}
	}
	return CartItem{}, false
}
