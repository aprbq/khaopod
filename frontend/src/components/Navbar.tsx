import { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { IconBag, IconClose, IconMenu, IconUser } from '@/components/icons'
import { Logo } from '@/components/Logo'
import { useAuth } from '@/features/auth/AuthContext'
import { useLang } from '@/i18n/LanguageContext'
import type { TranslationKey } from '@/i18n/translations'
import { cn } from '@/lib/utils'

interface NavItem {
  to: string
  labelKey: TranslationKey
  category?: string // ถ้ามี = กรองสินค้าตามหมวดนี้
}

const NAV: NavItem[] = [
  { to: '/', labelKey: 'nav.home' },
  { to: '/shop', labelKey: 'nav.collection' },
  { to: '/shop?category=sticker', labelKey: 'nav.stickers', category: 'sticker' },
]

function Brand({ className }: { className?: string }) {
  const { t } = useLang()
  return (
    <Link to="/" className={className} aria-label={t('nav.brandHome')}>
      <Logo className="size-9 rounded-sm" />
    </Link>
  )
}

function LangToggle() {
  const { lang, setLang } = useLang()
  return (
    <div className="flex items-center gap-1 text-xs font-bold">
      <button
        type="button"
        onClick={() => setLang('th')}
        className={cn('transition-colors', lang === 'th' ? 'text-foreground' : 'text-muted-foreground hover:text-foreground')}
      >
        TH
      </button>
      <span className="text-muted-foreground">/</span>
      <button
        type="button"
        onClick={() => setLang('en')}
        className={cn('transition-colors', lang === 'en' ? 'text-foreground' : 'text-muted-foreground hover:text-foreground')}
      >
        EN
      </button>
    </div>
  )
}

// ตรวจว่า nav item ไหน active (แยก /shop กับ /shop?category=... ให้ถูก)
function useIsActive() {
  const location = useLocation()
  const category = new URLSearchParams(location.search).get('category')
  return (item: NavItem): boolean => {
    if (item.to === '/') return location.pathname === '/'
    if (item.category) return location.pathname === '/shop' && category === item.category
    return location.pathname === '/shop' && !category
  }
}

export function Navbar() {
  const { status } = useAuth()
  const { t } = useLang()
  const [open, setOpen] = useState(false)
  const isActive = useIsActive()
  const accountTo = status === 'authenticated' ? '/account' : '/login'

  return (
    <header className="sticky top-0 z-40 border-b border-border bg-background/90 backdrop-blur">
      <div className="relative mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
        {/* ซ้าย: ปุ่มเมนู (มือถือ) + โลโก้ซ้าย (จอใหญ่) + เมนู */}
        <div className="flex items-center gap-3 md:gap-8">
          <button
            type="button"
            className="md:hidden"
            aria-label={t('nav.openMenu')}
            aria-expanded={open}
            onClick={() => setOpen((v) => !v)}
          >
            {open ? <IconClose className="size-5" /> : <IconMenu className="size-5" />}
          </button>

          {/* จอใหญ่: โลโก้อยู่ซ้าย */}
          <Brand className="hidden md:block" />

          <nav className="hidden items-center gap-8 md:flex">
            {NAV.map((n) => (
              <Link
                key={n.to}
                to={n.to}
                className={cn(
                  'text-xs font-bold uppercase tracking-widest transition-colors hover:text-foreground',
                  isActive(n) ? 'text-foreground' : 'text-muted-foreground',
                )}
              >
                {t(n.labelKey)}
              </Link>
            ))}
          </nav>
        </div>

        {/* จอมือถือ: โลโก้อยู่กึ่งกลางด้านบน */}
        <Brand className="absolute left-1/2 -translate-x-1/2 md:hidden" />

        {/* ขวา: สลับภาษา + บัญชี + ตะกร้า */}
        <div className="flex items-center gap-4">
          <LangToggle />
          <Link to={accountTo} aria-label={t('nav.account')} className="hover:text-accent">
            <IconUser className="size-5" />
          </Link>
          <Link to="/cart" aria-label={t('nav.cart')} className="relative hover:text-accent">
            <IconBag className="size-5" />
          </Link>
        </div>
      </div>

      {/* เมนูมือถือ */}
      {open && (
        <nav className="border-t border-border bg-background md:hidden">
          {NAV.map((n) => (
            <Link
              key={n.to}
              to={n.to}
              onClick={() => setOpen(false)}
              className={cn(
                'block border-b border-border px-4 py-3 text-sm font-semibold uppercase tracking-wide',
                isActive(n) ? 'text-foreground' : 'text-muted-foreground',
              )}
            >
              {t(n.labelKey)}
            </Link>
          ))}
        </nav>
      )}
    </header>
  )
}
