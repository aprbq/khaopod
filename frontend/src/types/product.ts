// สะท้อน response ของ /products ใน docs/rest_api.md (§5)

export interface PriceRange {
  min: number
  max: number
}

export interface Category {
  id: number
  name: string
  slug: string
}

export interface ProductVariant {
  id: number
  variant_name: string
  price: number
  stock_quantity: number
  sku?: string
}

export interface ProductImage {
  id: number
  url: string
  is_primary: boolean
  sort_order: number
}

// รายการสินค้า (§5.1) — ฟิลด์ย่อสำหรับการ์ด/กริด
export interface Product {
  id: number
  name: string
  slug: string
  base_price: number
  is_featured: boolean
  primary_image: string | null
  price_range: PriceRange
  in_stock: boolean
  category: Category | null // null = สินค้าไม่ถูกจัดหมวด
}

// รายละเอียดสินค้า (§5.2) — Product + description + รูปทั้งหมด + variants
export interface ProductDetail extends Product {
  description: string
  images: ProductImage[]
  variants: ProductVariant[]
}

// พารามิเตอร์ค้นหา/กรองของ GET /products
export interface ProductQuery {
  category?: string // slug ของหมวดหมู่
  search?: string
  sort?: string
  page?: number
  per_page?: number
}
