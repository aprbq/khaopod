import { useState } from 'react'
import { Link, NavLink } from 'react-router-dom'
import { IconBag, IconClose, IconMenu, IconUser } from '@/components/icons'
import { useAuth } from '@/features/auth/AuthContext'
import { cn } from '@/lib/utils'

const NAV = [
  { to: '/', label: 'หน้าแรก' },
  { to: '/shop', label: 'สินค้า' },
]

function Brand() {
  return (
    <Link to="/" className="flex items-baseline gap-1" aria-label="กองบัญชาการข่าวปด หน้าแรก">
      <span className="font-display text-lg tracking-tight">KHAOPOD</span>
      <span className="size-1.5 rounded-full bg-accent" />
    </Link>
  )
}

function navLinkClass({ isActive }: { isActive: boolean }): string {
  return cn(
    'text-xs font-bold uppercase tracking-widest transition-colors hover:text-foreground',
    isActive ? 'text-foreground' : 'text-muted-foreground',
  )
}

export function Navbar() {
  const { status } = useAuth()
  const [open, setOpen] = useState(false)
  const accountTo = status === 'authenticated' ? '/account' : '/login'

  return (
    <header className="sticky top-0 z-40 border-b border-border bg-background/90 backdrop-blur">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
        {/* ซ้าย: ปุ่มเมนู (มือถือ) + แบรนด์ */}
        <div className="flex items-center gap-3">
          <button
            type="button"
            className="md:hidden"
            aria-label="เปิดเมนู"
            aria-expanded={open}
            onClick={() => setOpen((v) => !v)}
          >
            {open ? <IconClose className="size-5" /> : <IconMenu className="size-5" />}
          </button>
          <Brand />
        </div>

        {/* กลาง: เมนู (จอใหญ่) */}
        <nav className="hidden items-center gap-8 md:flex">
          {NAV.map((n) => (
            <NavLink key={n.to} to={n.to} end={n.to === '/'} className={navLinkClass}>
              {n.label}
            </NavLink>
          ))}
        </nav>

        {/* ขวา: บัญชี + ตะกร้า */}
        <div className="flex items-center gap-4">
          <Link to={accountTo} aria-label="บัญชีของฉัน" className="hover:text-accent">
            <IconUser className="size-5" />
          </Link>
          <Link to="/cart" aria-label="ตะกร้าสินค้า" className="relative hover:text-accent">
            <IconBag className="size-5" />
          </Link>
        </div>
      </div>

      {/* เมนูมือถือ */}
      {open && (
        <nav className="border-t border-border bg-background md:hidden">
          {NAV.map((n) => (
            <NavLink
              key={n.to}
              to={n.to}
              end={n.to === '/'}
              onClick={() => setOpen(false)}
              className="block border-b border-border px-4 py-3 text-sm font-semibold uppercase tracking-wide"
            >
              {n.label}
            </NavLink>
          ))}
        </nav>
      )}
    </header>
  )
}
