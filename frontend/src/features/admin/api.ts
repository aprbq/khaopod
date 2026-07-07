import { apiFetch } from '@/lib/apiClient'
import type {
  AdminOrder,
  AdminProduct,
  AdminSummary,
  AdminUser,
  ProductInput,
  VariantInput,
} from '@/types/admin'
import type { OrderStatus, PaymentStatus } from '@/types/order'
import type { Category } from '@/types/product'

// รวม call ของหลังบ้านแอดมินไว้ที่เดียว (ดู docs/rest_api.md §11) — ต้อง role admin
export const adminApi = {
  summary: () => apiFetch<AdminSummary>('/admin/dashboard/summary'),

  listOrders: (status?: OrderStatus) =>
    apiFetch<AdminOrder[]>(`/admin/orders${status ? `?status=${status}` : ''}`),

  getOrder: (orderNumber: string) =>
    apiFetch<AdminOrder>(`/admin/orders/${encodeURIComponent(orderNumber)}`),

  updateStatus: (orderNumber: string, status: OrderStatus, note?: string) =>
    apiFetch<AdminOrder>(`/admin/orders/${encodeURIComponent(orderNumber)}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status, note: note ?? '' }),
    }),

  verifyPayment: (paymentId: number, status: Extract<PaymentStatus, 'paid' | 'failed'>) =>
    apiFetch<AdminOrder>(`/admin/payments/${paymentId}/verify`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    }),

  // ---- Users ----

  listUsers: () => apiFetch<AdminUser[]>('/admin/users'),

  // ---- Catalog (§5.3) ----

  listProducts: () => apiFetch<AdminProduct[]>('/admin/products'),

  getProduct: (id: number) => apiFetch<AdminProduct>(`/admin/products/${id}`),

  createProduct: (input: ProductInput) =>
    apiFetch<AdminProduct>('/admin/products', { method: 'POST', body: JSON.stringify(input) }),

  updateProduct: (id: number, input: ProductInput) =>
    apiFetch<AdminProduct>(`/admin/products/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(input),
    }),

  deleteProduct: (id: number) => apiFetch<void>(`/admin/products/${id}`, { method: 'DELETE' }),

  createVariant: (productId: number, input: VariantInput) =>
    apiFetch<unknown>(`/admin/products/${productId}/variants`, {
      method: 'POST',
      body: JSON.stringify(input),
    }),

  updateVariant: (variantId: number, input: VariantInput) =>
    apiFetch<unknown>(`/admin/variants/${variantId}`, {
      method: 'PATCH',
      body: JSON.stringify(input),
    }),

  deleteVariant: (variantId: number) =>
    apiFetch<void>(`/admin/variants/${variantId}`, { method: 'DELETE' }),

  addProductImage: (productId: number, file: File) => {
    const form = new FormData()
    form.append('image', file)
    return apiFetch<unknown>(`/admin/products/${productId}/images`, {
      method: 'POST',
      body: form,
    })
  },

  deleteProductImage: (imageId: number) =>
    apiFetch<void>(`/admin/images/${imageId}`, { method: 'DELETE' }),

  setPrimaryImage: (productId: number, imageId: number) =>
    apiFetch<unknown>(`/admin/products/${productId}/images/${imageId}/primary`, {
      method: 'POST',
    }),
}

// หมวดหมู่ (endpoint สาธารณะ — ใช้ใน select ของฟอร์มสินค้า)
export const categoryApi = {
  list: () => apiFetch<Category[]>('/categories'),
}
