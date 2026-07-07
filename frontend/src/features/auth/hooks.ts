import { useMutation } from '@tanstack/react-query'
import type { UpdateProfileInput } from '@/types/api'
import { authApi } from './api'

export function useRequestOtp() {
  return useMutation({ mutationFn: (email: string) => authApi.requestOtp(email) })
}

export function useVerifyOtp() {
  return useMutation({
    mutationFn: (vars: { email: string; code: string }) => authApi.verifyOtp(vars.email, vars.code),
  })
}

export function useUpdateProfile() {
  return useMutation({ mutationFn: (input: UpdateProfileInput) => authApi.updateProfile(input) })
}

export function useUploadAvatar() {
  return useMutation({ mutationFn: (file: File) => authApi.uploadAvatar(file) })
}
