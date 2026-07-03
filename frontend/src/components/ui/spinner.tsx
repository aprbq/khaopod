import { cn } from '@/lib/utils'

// spinner เล็ก ๆ ใช้ตอนรอ mutation/query (กัน double-submit + บอกสถานะ loading)
export function Spinner({ className }: { className?: string }) {
  return (
    <svg
      className={cn('animate-spin', className)}
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      role="status"
      aria-label="กำลังโหลด"
    >
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeOpacity="0.25" strokeWidth="4" />
      <path d="M22 12a10 10 0 0 1-10 10" stroke="currentColor" strokeWidth="4" strokeLinecap="round" />
    </svg>
  )
}
