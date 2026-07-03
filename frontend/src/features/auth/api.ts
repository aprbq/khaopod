import { apiFetch } from '@/lib/apiClient'
import type { AuthResult, OTPChallenge, UpdateProfileInput, User } from '@/types/api'

// รวม call ของ auth/user ไว้ที่เดียว (ดู docs/rest_api.md §2, §3)
export const authApi = {
  requestOtp: (email: string) =>
    apiFetch<OTPChallenge>('/auth/otp/request', {
      method: 'POST',
      body: JSON.stringify({ email }),
    }),

  verifyOtp: (email: string, code: string) =>
    apiFetch<AuthResult>('/auth/otp/verify', {
      method: 'POST',
      body: JSON.stringify({ email, code }),
    }),

  me: () => apiFetch<User>('/auth/me'),

  logout: () => apiFetch<void>('/auth/logout', { method: 'POST' }),

  updateProfile: (input: UpdateProfileInput) =>
    apiFetch<User>('/me', { method: 'PATCH', body: JSON.stringify(input) }),
}
