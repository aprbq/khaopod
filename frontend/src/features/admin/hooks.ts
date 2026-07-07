import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/features/auth/AuthContext'
import type { ProductInput, VariantInput } from '@/types/admin'
import type { OrderStatus, PaymentStatus } from '@/types/order'
import { adminApi, categoryApi } from './api'

const ADMIN_KEY = ['admin']

// query ฝั่งแอดมินยิงเฉพาะเมื่อ role เป็น admin — กันยิงแล้วโดน 403 ตอน guard ยังไม่ทัน redirect
function useIsAdmin(): boolean {
  const { user, status } = useAuth()
  return status === 'authenticated' && user?.role === 'admin'
}

export function useAdminSummary() {
  const enabled = useIsAdmin()
  return useQuery({
    queryKey: [...ADMIN_KEY, 'summary'],
    queryFn: adminApi.summary,
    enabled,
  })
}

export function useAdminOrders(status?: OrderStatus) {
  const enabled = useIsAdmin()
  return useQuery({
    queryKey: [...ADMIN_KEY, 'orders', status ?? 'all'],
    queryFn: () => adminApi.listOrders(status),
    enabled,
  })
}

export function useAdminOrder(orderNumber: string) {
  const enabled = useIsAdmin()
  return useQuery({
    queryKey: [...ADMIN_KEY, 'order', orderNumber],
    queryFn: () => adminApi.getOrder(orderNumber),
    enabled: enabled && orderNumber !== '',
  })
}

// mutation แล้ว invalidate ทั้งฝั่งแอดมิน (list/summary/detail) ให้ sync
function useAdminMutation<TArgs>(fn: (args: TArgs) => Promise<unknown>) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: fn,
    onSuccess: () => qc.invalidateQueries({ queryKey: ADMIN_KEY }),
  })
}

export function useAdminUpdateStatus() {
  return useAdminMutation(
    ({ orderNumber, status, note }: { orderNumber: string; status: OrderStatus; note?: string }) =>
      adminApi.updateStatus(orderNumber, status, note),
  )
}

export function useAdminVerifyPayment() {
  return useAdminMutation(
    ({
      paymentId,
      status,
    }: {
      paymentId: number
      status: Extract<PaymentStatus, 'paid' | 'failed'>
    }) => adminApi.verifyPayment(paymentId, status),
  )
}

// ---- Users ----

export function useAdminUsers() {
  const enabled = useIsAdmin()
  return useQuery({
    queryKey: [...ADMIN_KEY, 'users'],
    queryFn: adminApi.listUsers,
    enabled,
  })
}

// ---- Catalog ----

export function useAdminProducts() {
  const enabled = useIsAdmin()
  return useQuery({
    queryKey: [...ADMIN_KEY, 'products'],
    queryFn: adminApi.listProducts,
    enabled,
  })
}

export function useAdminProduct(id: number) {
  const enabled = useIsAdmin()
  return useQuery({
    queryKey: [...ADMIN_KEY, 'product', id],
    queryFn: () => adminApi.getProduct(id),
    enabled: enabled && id > 0,
  })
}

// mutation ของแคตตาล็อกต้องล้าง cache หน้าร้านด้วย — สินค้าที่แก้ต้องสะท้อนทันทีทั้งสองฝั่ง
function useCatalogMutation<TArgs, TOut>(fn: (args: TArgs) => Promise<TOut>) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: fn,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ADMIN_KEY })
      qc.invalidateQueries({ queryKey: ['products'] })
    },
  })
}

export function useAdminCreateProduct() {
  return useCatalogMutation((input: ProductInput) => adminApi.createProduct(input))
}

export function useAdminUpdateProduct() {
  return useCatalogMutation(({ id, input }: { id: number; input: ProductInput }) =>
    adminApi.updateProduct(id, input),
  )
}

export function useAdminDeleteProduct() {
  return useCatalogMutation((id: number) => adminApi.deleteProduct(id))
}

export function useAdminCreateVariant() {
  return useCatalogMutation(({ productId, input }: { productId: number; input: VariantInput }) =>
    adminApi.createVariant(productId, input),
  )
}

export function useAdminUpdateVariant() {
  return useCatalogMutation(({ variantId, input }: { variantId: number; input: VariantInput }) =>
    adminApi.updateVariant(variantId, input),
  )
}

export function useAdminDeleteVariant() {
  return useCatalogMutation((variantId: number) => adminApi.deleteVariant(variantId))
}

export function useAdminAddImage() {
  return useCatalogMutation(({ productId, file }: { productId: number; file: File }) =>
    adminApi.addProductImage(productId, file),
  )
}

export function useAdminDeleteImage() {
  return useCatalogMutation((imageId: number) => adminApi.deleteProductImage(imageId))
}

export function useAdminSetPrimaryImage() {
  return useCatalogMutation(({ productId, imageId }: { productId: number; imageId: number }) =>
    adminApi.setPrimaryImage(productId, imageId),
  )
}

// หมวดหมู่ (สาธารณะ) — cache นานได้ เปลี่ยนไม่บ่อย
export function useCategories() {
  return useQuery({ queryKey: ['categories'], queryFn: categoryApi.list, staleTime: 5 * 60_000 })
}
