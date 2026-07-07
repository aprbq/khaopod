// สะท้อน response ของ /orders ใน docs/rest_api.md (§9, §10)

export type OrderStatus =
  | 'pending'
  | 'paid'
  | 'preparing'
  | 'shipped'
  | 'delivered'
  | 'completed'
  | 'cancelled'
  | 'refunded'

export type PaymentStatus = 'unpaid' | 'pending_review' | 'paid' | 'failed' | 'refunded'

// ระบบรองรับสองวิธีตอนนี้ (ดู §9.1)
export type PaymentMethod = 'promptpay' | 'bank_transfer'

export interface OrderItem {
  product_name: string
  variant_name: string
  unit_price: number
  quantity: number
  line_total: number
}

export interface ShippingAddress {
  recipient: string
  phone: string
  address: string
  subdistrict: string
  district: string
  province: string
  postal_code: string
}

export interface Payment {
  id: number
  method: PaymentMethod
  amount: number
  status: PaymentStatus
  slip_url?: string
  transaction_ref?: string
  created_at: string
}

export interface Order {
  order_number: string
  status: OrderStatus
  payment_status: PaymentStatus
  payment_method?: PaymentMethod
  subtotal: number
  discount_amount: number
  shipping_fee: number
  total_amount: number
  items: OrderItem[]
  shipping_address: ShippingAddress
  customer_note?: string
  payment?: Payment
  placed_at: string
}

export interface PlaceOrderInput {
  address_id: number
  payment_method: PaymentMethod
  customer_note?: string
}

export interface SubmitPaymentInput {
  method: PaymentMethod
  amount: number
  transaction_ref?: string
  slip: File
}
