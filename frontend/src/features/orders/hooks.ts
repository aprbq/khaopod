import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/features/auth/AuthContext'
import type { PlaceOrderInput, SubmitPaymentInput } from '@/types/order'
import { orderApi } from './api'

const ORDERS_KEY = ['orders']
const CART_KEY = ['cart']

export function useOrders() {
  const { status } = useAuth()
  return useQuery({
    queryKey: ORDERS_KEY,
    queryFn: orderApi.list,
    enabled: status === 'authenticated',
  })
}

export function useOrder(orderNumber: string) {
  const { status } = useAuth()
  return useQuery({
    queryKey: [...ORDERS_KEY, orderNumber],
    queryFn: () => orderApi.get(orderNumber),
    enabled: status === 'authenticated' && orderNumber !== '',
  })
}

export function usePlaceOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: PlaceOrderInput) => orderApi.place(input),
    onSuccess: (order) => {
      // checkout ปิดตะกร้าฝั่ง server แล้ว — ล้าง cache ตะกร้า/ออเดอร์ให้ sync ทันที
      qc.invalidateQueries({ queryKey: CART_KEY })
      qc.invalidateQueries({ queryKey: ORDERS_KEY })
      qc.setQueryData([...ORDERS_KEY, order.order_number], order)
    },
  })
}

export function useCancelOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ orderNumber, reason }: { orderNumber: string; reason?: string }) =>
      orderApi.cancel(orderNumber, reason),
    onSuccess: (order) => {
      qc.setQueryData([...ORDERS_KEY, order.order_number], order)
      qc.invalidateQueries({ queryKey: ORDERS_KEY, exact: true })
    },
  })
}

export function useSubmitPayment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ orderNumber, input }: { orderNumber: string; input: SubmitPaymentInput }) =>
      orderApi.submitPayment(orderNumber, input),
    onSuccess: (order) => {
      qc.setQueryData([...ORDERS_KEY, order.order_number], order)
      qc.invalidateQueries({ queryKey: ORDERS_KEY, exact: true })
    },
  })
}
