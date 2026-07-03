// สะท้อน response ของ /products ใน docs/rest_api.md (§5)
// (ตอนนี้ยังใช้ mock — เมื่อ backend มี endpoint สินค้าค่อยสลับมา apiFetch)

export interface PriceRange {
  min: number
  max: number
}

export interface ProductVariant {
  id: number
  variant_name: string
  price: number
  stock_quantity: number
  sku?: string
}

export interface Product {
  id: number
  name: string
  slug: string
  base_price: number
  is_featured: boolean
  primary_image: string | null
  price_range: PriceRange
  in_stock: boolean
  category: string // label หมวดหมู่ (ไว้โชว์)
}

export interface ProductDetail extends Product {
  description: string
  images: string[]
  variants: ProductVariant[]
}
