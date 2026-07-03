import { Navigate, Outlet } from 'react-router-dom'
import { Spinner } from '@/components/ui/spinner'
import { useAuth } from './AuthContext'

// ครอบ route ที่ต้องล็อกอิน — ยังไม่รู้สถานะ = รอ, เป็น guest = เด้งไป /login
export function ProtectedRoute() {
  const { status } = useAuth()

  if (status === 'loading') {
    return (
      <div className="grid min-h-screen place-items-center">
        <Spinner className="size-6 text-muted-foreground" />
      </div>
    )
  }
  if (status === 'unauthenticated') {
    return <Navigate to="/login" replace />
  }
  return <Outlet />
}
