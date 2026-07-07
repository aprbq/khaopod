import { apiFetch } from '@/lib/apiClient'
import type { Order, PlaceOrderInput, SubmitPaymentInput } from '@/types/order'

// รวม call ของคำสั่งซื้อ + ชำระเงินไว้ที่เดียว (ดู docs/rest_api.md §9, §10)
export const orderApi = {
  place: (input: PlaceOrderInput) =>
    apiFetch<Order>('/orders', { method: 'POST', body: JSON.stringify(input) }),

  list: () => apiFetch<Order[]>('/orders'),

  get: (orderNumber: string) => apiFetch<Order>(`/orders/${encodeURIComponent(orderNumber)}`),

  cancel: (orderNumber: string, reason?: string) =>
    apiFetch<Order>(`/orders/${encodeURIComponent(orderNumber)}/cancel`, {
      method: 'POST',
      body: JSON.stringify({ reason: reason ?? '' }),
    }),

  submitPayment: (orderNumber: string, input: SubmitPaymentInput) => {
    const form = new FormData()
    form.append('method', input.method)
    form.append('amount', String(input.amount))
    if (input.transaction_ref) form.append('transaction_ref', input.transaction_ref)
    form.append('slip', input.slip)
    return apiFetch<Order>(`/orders/${encodeURIComponent(orderNumber)}/payment`, {
      method: 'POST',
      body: form,
    })
  },
}
