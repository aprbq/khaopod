// สะท้อน response ของ /admin/* ใน docs/rest_api.md (§5.3, §11)

import type { Order } from './order'
import type { ProductDetail } from './product'
import type { Role } from './api'

// ออเดอร์ในมุมแอดมิน = ออเดอร์ปกติ + อีเมลเจ้าของ
export interface AdminOrder extends Order {
  user_email: string
}

export interface AdminSummary {
  orders_total: number
  orders_pending: number
  payments_pending_review: number
  revenue_paid: number
}

// สินค้าในมุมแอดมิน = รายละเอียดเต็ม + สถานะเปิด/ปิดขาย (หน้าร้านไม่โชว์ฟิลด์นี้)
export interface AdminProduct extends ProductDetail {
  is_active: boolean
}

export interface ProductInput {
  name: string
  slug: string
  description?: string
  base_price: number
  category_id: number | null
  is_active: boolean
  is_featured: boolean
}

export interface VariantInput {
  variant_name: string
  color?: string
  sku?: string
  price: number
  stock_quantity: number
  is_active: boolean
}

export interface AdminUser {
  public_id: string
  email: string
  display_name: string
  avatar_url: string
  phone?: string
  role: Role
  is_active: boolean
  last_login_at: string | null
  created_at: string
}
