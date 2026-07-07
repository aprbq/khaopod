// type ที่สะท้อน response ของ API — ต้องตรงกับ docs/rest_api.md

export type Role = 'customer' | 'admin'

export interface User {
  public_id: string
  email: string
  display_name: string
  avatar_url: string
  phone?: string
  role: Role
  created_at: string
  updated_at: string
}

// ผลหลังขอ OTP (ยังไม่ล็อกอิน)
export interface OTPChallenge {
  email: string
  display_name?: string
  expires_in: number
  message: string
}

// ผลหลัง verify OTP สำเร็จ
export interface AuthResult {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
  user: User
}

export interface UpdateProfileInput {
  display_name?: string
  phone?: string
}
