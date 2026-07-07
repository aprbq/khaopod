// API client ตัวเดียวของทั้งเว็บ — จัดการ envelope + แนบ token + refresh อัตโนมัติ
// อย่า fetch ดิบ ๆ รายจุด ให้เรียกผ่าน apiFetch เสมอ

// ยิงตรงไป backend ตาม VITE_API_BASE_URL ใน .env (เช่น http://localhost:8080/api/v1)
// ยิงข้าม origin → backend เปิด CORS ให้ origin นี้แล้ว
const BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

// access token เก็บใน memory เท่านั้น (ไม่ลง localStorage — ปลอดภัยกว่า)
let accessToken: string | null = null
export function setAccessToken(token: string | null): void {
  accessToken = token
}
export function getAccessToken(): string | null {
  return accessToken
}

// เรียกเมื่อ session หมดจริง (refresh ไม่ผ่าน) — ให้ AuthProvider พาไปหน้า login
let onSessionExpired: (() => void) | null = null
export function setOnSessionExpired(cb: (() => void) | null): void {
  onSessionExpired = cb
}

// error ที่พก code (machine-readable) + message (ภาษาไทยพร้อมโชว์ผู้ใช้)
export class ApiError extends Error {
  code: string
  status: number
  constructor(code: string, message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }
}

interface Envelope<T> {
  success: boolean
  data?: T
  error?: { code: string; message: string }
}

function buildRequest(path: string, opts?: RequestInit): Request {
  const headers = new Headers(opts?.headers)
  // FormData ห้ามตั้ง Content-Type เอง — browser ต้องเป็นคนใส่ boundary ของ multipart
  if (opts?.body && !(opts.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`)
  }
  return new Request(`${BASE_URL}${path}`, { ...opts, headers, credentials: 'include' })
}

// ขอ access token ใหม่จาก refresh cookie (httpOnly) — คืน true ถ้าสำเร็จ
// ไม่ throw แม้ backend ล่ม (หน้าร้าน public ต้องเปิดได้แม้ไม่มี backend)
export async function tryRefresh(): Promise<boolean> {
  try {
    const res = await fetch(`${BASE_URL}/auth/refresh`, { method: 'POST', credentials: 'include' })
    if (!res.ok) return false
    const body = (await res.json().catch(() => null)) as Envelope<{ access_token: string }> | null
    if (!body?.success || !body.data) return false
    accessToken = body.data.access_token
    return true
  } catch {
    return false
  }
}

export async function apiFetch<T>(path: string, opts?: RequestInit): Promise<T> {
  let res = await fetch(buildRequest(path, opts))

  // 401 ทั้งที่มี token → ลอง refresh 1 ครั้งแล้ว retry
  if (res.status === 401 && accessToken) {
    if (await tryRefresh()) {
      res = await fetch(buildRequest(path, opts))
    } else {
      accessToken = null
      onSessionExpired?.()
    }
  }

  if (res.status === 204) {
    return undefined as T
  }

  const body = (await res.json().catch(() => null)) as Envelope<T> | null
  if (!body || typeof body.success !== 'boolean') {
    throw new ApiError('NETWORK_ERROR', 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้ กรุณาลองใหม่', res.status)
  }
  if (!body.success) {
    const err = body.error ?? { code: 'UNKNOWN', message: 'เกิดข้อผิดพลาด' }
    throw new ApiError(err.code, err.message, res.status)
  }
  return body.data as T
}
