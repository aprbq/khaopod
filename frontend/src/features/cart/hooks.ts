import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/features/auth/AuthContext'
import type { Cart } from '@/types/cart'
import { cartApi } from './api'

const CART_KEY = ['cart']

const emptyCart: Cart = { id: 0, items: [], subtotal: 0, item_count: 0 }

// ดึงตะกร้า — ยิงเฉพาะตอนล็อกอินแล้ว (endpoint ต้องมี auth)
export function useCart() {
  const { status } = useAuth()
  return useQuery({
    queryKey: CART_KEY,
    queryFn: cartApi.get,
    enabled: status === 'authenticated',
  })
}

// mutation ที่ backend คืน Cart ใหม่ → set ลง cache ตรง ๆ (ไม่ต้อง refetch)
function useCartMutation<TArgs>(fn: (args: TArgs) => Promise<Cart>) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: fn,
    onSuccess: (cart) => qc.setQueryData(CART_KEY, cart),
  })
}

export function useAddToCart() {
  return useCartMutation(({ variantId, quantity }: { variantId: number; quantity: number }) =>
    cartApi.addItem(variantId, quantity),
  )
}

export function useUpdateCartItem() {
  return useCartMutation(({ itemId, quantity }: { itemId: number; quantity: number }) =>
    cartApi.updateItem(itemId, quantity),
  )
}

export function useRemoveCartItem() {
  return useCartMutation(({ itemId }: { itemId: number }) => cartApi.removeItem(itemId))
}

export function useClearCart() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: cartApi.clear,
    onSuccess: () => qc.setQueryData(CART_KEY, emptyCart),
  })
}
