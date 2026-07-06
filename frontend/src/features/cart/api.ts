import { apiFetch } from '@/lib/apiClient'
import type { Cart } from '@/types/cart'

// รวม call ของตะกร้าไว้ที่เดียว (ดู docs/rest_api.md §6) — ทุก endpoint ต้องล็อกอิน
// op ที่แก้ตะกร้าคืน Cart ล่าสุดกลับมา จึงเอาไปอัปเดต cache ได้เลย
export const cartApi = {
  get: () => apiFetch<Cart>('/cart'),

  addItem: (variantId: number, quantity: number) =>
    apiFetch<Cart>('/cart/items', {
      method: 'POST',
      body: JSON.stringify({ variant_id: variantId, quantity }),
    }),

  updateItem: (itemId: number, quantity: number) =>
    apiFetch<Cart>(`/cart/items/${itemId}`, {
      method: 'PATCH',
      body: JSON.stringify({ quantity }),
    }),

  removeItem: (itemId: number) => apiFetch<Cart>(`/cart/items/${itemId}`, { method: 'DELETE' }),

  clear: () => apiFetch<void>('/cart', { method: 'DELETE' }),
}
