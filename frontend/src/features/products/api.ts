import { apiFetch } from '@/lib/apiClient'
import type { Product, ProductDetail, ProductQuery } from '@/types/product'

// สร้าง query string จากพารามิเตอร์ที่มีค่าเท่านั้น (ข้ามค่าว่าง/undefined)
function toQueryString(q: ProductQuery = {}): string {
  const params = new URLSearchParams()
  if (q.category) params.set('category', q.category)
  if (q.search) params.set('search', q.search)
  if (q.sort) params.set('sort', q.sort)
  if (q.page) params.set('page', String(q.page))
  if (q.per_page) params.set('per_page', String(q.per_page))
  const s = params.toString()
  return s ? `?${s}` : ''
}

// รวม call ของแคตตาล็อกสินค้าไว้ที่เดียว (ดู docs/rest_api.md §5)
// apiFetch แกะ envelope ให้แล้ว: list คืน data (array), detail คืน object
export const productApi = {
  list: (query?: ProductQuery) => apiFetch<Product[]>(`/products${toQueryString(query)}`),

  get: (slug: string) => apiFetch<ProductDetail>(`/products/${encodeURIComponent(slug)}`),
}
