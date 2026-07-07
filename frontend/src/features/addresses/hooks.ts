import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/features/auth/AuthContext'
import type { AddressInput } from '@/types/address'
import { addressApi } from './api'

const ADDRESSES_KEY = ['addresses']

export function useAddresses() {
  const { status } = useAuth()
  return useQuery({
    queryKey: ADDRESSES_KEY,
    queryFn: addressApi.list,
    enabled: status === 'authenticated',
  })
}

// mutation ทุกตัว invalidate ทั้งชุด — ที่อยู่มีไม่กี่รายการ refetch ถูกกว่า sync cache เอง
function useAddressMutation<TArgs, TOut>(fn: (args: TArgs) => Promise<TOut>) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: fn,
    onSuccess: () => qc.invalidateQueries({ queryKey: ADDRESSES_KEY }),
  })
}

export function useCreateAddress() {
  return useAddressMutation((input: AddressInput) => addressApi.create(input))
}

export function useUpdateAddress() {
  return useAddressMutation(({ id, input }: { id: number; input: AddressInput }) =>
    addressApi.update(id, input),
  )
}

export function useDeleteAddress() {
  return useAddressMutation((id: number) => addressApi.remove(id))
}

export function useSetDefaultAddress() {
  return useAddressMutation((id: number) => addressApi.setDefault(id))
}
