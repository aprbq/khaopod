import type { ReactNode } from 'react'

// เลย์เอาต์การ์ดกลางจอ ใช้ร่วมกันหน้า login / verify
export function AuthShell({ children }: { children: ReactNode }) {
  return (
    <div className="grid min-h-screen place-items-center bg-muted/40 px-4 py-10">
      <div className="w-full max-w-sm">
        <div className="mb-6 text-center">
          <p className="text-xs font-bold uppercase tracking-[0.2em] text-accent">
            Khaopod News Shop
          </p>
          <h1 className="mt-1 text-2xl font-bold tracking-tight">กองบัญชาการข่าวปด</h1>
        </div>
        {children}
      </div>
    </div>
  )
}
