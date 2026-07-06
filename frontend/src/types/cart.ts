// สะท้อน response ของ /cart ใน docs/rest_api.md (§6)

export interface CartItem {
  id: number
  variant_id: number
  product_name: string
  variant_name: string // ไซซ์
  color?: string // สี
  unit_price: number
  quantity: number
  line_total: number
  image?: string
  in_stock: boolean // สต็อกยังพอกับจำนวนในตะกร้า
}

export interface Cart {
  id: number
  items: CartItem[]
  subtotal: number
  item_count: number
}
