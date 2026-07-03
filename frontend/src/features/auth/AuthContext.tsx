import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { setAccessToken, setOnSessionExpired, tryRefresh } from '@/lib/apiClient'
import type { AuthResult, User } from '@/types/api'
import { authApi } from './api'

type Status = 'loading' | 'authenticated' | 'unauthenticated'

interface AuthContextValue {
  user: User | null
  status: Status
  setSession: (result: AuthResult) => void
  setUser: (user: User) => void
  signOut: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [status, setStatus] = useState<Status>('loading')

  useEffect(() => {
    let active = true

    // เมื่อ session หมดจริง (refresh ไม่ผ่านระหว่างเรียก API) → เด้งกลับเป็น guest
    setOnSessionExpired(() => {
      if (!active) return
      setUser(null)
      setStatus('unauthenticated')
    })

    // boot: ลอง refresh จาก httpOnly cookie → ถ้าได้ ดึงโปรไฟล์มาแสดง (คงล็อกอินข้าม reload)
    void (async () => {
      const ok = await tryRefresh()
      if (!active) return
      if (!ok) {
        setStatus('unauthenticated')
        return
      }
      try {
        const me = await authApi.me()
        if (!active) return
        setUser(me)
        setStatus('authenticated')
      } catch {
        if (active) setStatus('unauthenticated')
      }
    })()

    return () => {
      active = false
      setOnSessionExpired(null)
    }
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      status,
      setSession: (result) => {
        setAccessToken(result.access_token)
        setUser(result.user)
        setStatus('authenticated')
      },
      setUser: (u) => setUser(u),
      signOut: async () => {
        try {
          await authApi.logout()
        } catch {
          // เพิกเฉย error ตอน logout — ยังไงก็เคลียร์ session ฝั่ง client
        }
        setAccessToken(null)
        setUser(null)
        setStatus('unauthenticated')
      },
    }),
    [user, status],
  )

  return <AuthContext value={value}>{children}</AuthContext>
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth ต้องอยู่ภายใต้ <AuthProvider>')
  return ctx
}
