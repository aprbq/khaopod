import { useState } from 'react'
import { Link, NavLink, Outlet } from 'react-router-dom'
import { IconBag, IconClose, IconMenu } from '@/components/icons'
import { useLang } from '@/i18n/LanguageContext'
import { cn } from '@/lib/utils'

// เชลล์หลังบ้านแอดมิน — คนละโลกกับหน้าร้าน: sidebar ดำด้านซ้าย + พื้นที่ทำงานโทนเทา
// จอเล็ก sidebar ยุบเป็น drawer เปิดจากปุ่มบน top bar
export function AdminLayout() {
  const { t } = useLang()
  const [open, setOpen] = useState(false)

  const nav = (
    <nav className="flex flex-1 flex-col gap-1 p-3">
      {[
        { to: '/admin', label: t('admin.dashboard'), end: true },
        { to: '/admin/orders', label: t('admin.orders'), end: false },
        { to: '/admin/products', label: t('admin.products'), end: false },
        { to: '/admin/users', label: t('admin.users'), end: false },
      ].map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          end={item.end}
          onClick={() => setOpen(false)}
          className={({ isActive }) =>
            cn(
              'rounded-md px-3 py-2 text-sm font-semibold uppercase tracking-wide transition-colors',
              isActive
                ? 'bg-accent text-accent-foreground'
                : 'text-primary-foreground/70 hover:bg-primary-foreground/10 hover:text-primary-foreground',
            )
          }
        >
          {item.label}
        </NavLink>
      ))}
    </nav>
  )

  const sidebarInner = (
    <>
      <div className="flex items-center gap-2 border-b border-primary-foreground/10 px-4 py-5">
        <IconBag className="size-5" />
        <span className="font-display text-lg uppercase leading-none">{t('admin.brand')}</span>
      </div>
      {nav}
      <Link
        to="/"
        className="border-t border-primary-foreground/10 px-4 py-4 text-sm text-primary-foreground/70 transition-colors hover:text-primary-foreground"
      >
        ← {t('admin.backToStore')}
      </Link>
    </>
  )

  return (
    <div className="flex min-h-screen bg-muted/40 text-foreground">
      {/* sidebar จอใหญ่ */}
      <aside className="sticky top-0 hidden h-screen w-60 flex-col bg-primary text-primary-foreground lg:flex">
        {sidebarInner}
      </aside>

      {/* drawer จอเล็ก */}
      {open && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <button
            type="button"
            aria-label={t('common.close')}
            className="absolute inset-0 bg-black/50"
            onClick={() => setOpen(false)}
          />
          <aside className="absolute inset-y-0 left-0 flex w-60 flex-col bg-primary text-primary-foreground">
            <button
              type="button"
              aria-label={t('common.close')}
              className="absolute right-3 top-4 p-1"
              onClick={() => setOpen(false)}
            >
              <IconClose className="size-5" />
            </button>
            {sidebarInner}
          </aside>
        </div>
      )}

      <div className="flex min-w-0 flex-1 flex-col">
        {/* top bar จอเล็ก */}
        <header className="flex items-center gap-3 border-b border-border bg-background px-4 py-3 lg:hidden">
          <button type="button" aria-label={t('nav.openMenu')} onClick={() => setOpen(true)}>
            <IconMenu className="size-6" />
          </button>
          <span className="font-display text-lg uppercase">{t('admin.brand')}</span>
        </header>

        <main className="mx-auto w-full max-w-5xl flex-1 px-4 py-6 md:px-8 md:py-8">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
