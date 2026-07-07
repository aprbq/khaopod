import { apiFetch } from '@/lib/apiClient'
import type { Address, AddressInput } from '@/types/address'

// รวม call ของที่อยู่จัดส่งไว้ที่เดียว (ดู docs/rest_api.md §7) — ทุก endpoint ต้องล็อกอิน
export const addressApi = {
  list: () => apiFetch<Address[]>('/addresses'),

  create: (input: AddressInput) =>
    apiFetch<Address>('/addresses', { method: 'POST', body: JSON.stringify(input) }),

  update: (id: number, input: AddressInput) =>
    apiFetch<Address>(`/addresses/${id}`, { method: 'PATCH', body: JSON.stringify(input) }),

  remove: (id: number) => apiFetch<void>(`/addresses/${id}`, { method: 'DELETE' }),

  setDefault: (id: number) => apiFetch<Address>(`/addresses/${id}/default`, { method: 'POST' }),
}
