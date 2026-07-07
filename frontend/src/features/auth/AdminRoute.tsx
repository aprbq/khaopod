import { Navigate, Outlet } from 'react-router-dom'
import { Spinner } from '@/components/ui/spinner'
import { useAuth } from './AuthContext'

// ครอบ route หลังบ้าน — ต้องล็อกอิน "และ" เป็นแอดมิน (ลูกค้าธรรมดาเด้งกลับหน้าร้าน)
// backend ก็เช็ค role ซ้ำที่ middleware อยู่แล้ว — guard นี้เป็นแค่ UX
export function AdminRoute() {
  const { user, status } = useAuth()

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
  if (user?.role !== 'admin') {
    return <Navigate to="/" replace />
  }
  return <Outlet />
}
