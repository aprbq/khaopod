import logoUrl from '@/asset/logo_khoapod.jpg'
import { cn } from '@/lib/utils'

// โลโก้ร้าน — ใช้ซ้ำในนาวบาร์ / footer / หน้า auth
export function Logo({ className }: { className?: string }) {
  return (
    <img
      src={logoUrl}
      alt="กองบัญชาการข่าวปด"
      className={cn('block object-cover', className)}
    />
  )
}
