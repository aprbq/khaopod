import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { Logo } from '@/components/Logo'
import { useLang } from '@/i18n/LanguageContext'

// เลย์เอาต์การ์ดกลางจอ ใช้ร่วมกันหน้า login / verify
export function AuthShell({ children }: { children: ReactNode }) {
  const { t } = useLang()
  return (
    <div className="grid min-h-screen place-items-center bg-muted/40 px-4 py-10">
      <div className="w-full max-w-sm">
        <div className="mb-6 flex flex-col items-center text-center">
          <Link to="/" aria-label={t('nav.brandHome')}>
            <Logo className="size-16 rounded-md" />
          </Link>
          <h1 className="mt-3 text-2xl font-bold tracking-tight">{t('login.brand')}</h1>
        </div>
        {children}
      </div>
    </div>
  )
}
