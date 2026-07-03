import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { Logo } from '@/components/Logo'
import { useLang } from '@/i18n/LanguageContext'

export function Footer() {
  const { t } = useLang()
  return (
    <footer className="border-t border-border">
      <div className="mx-auto max-w-6xl px-4 py-12">
        <div className="grid grid-cols-2 gap-8 md:grid-cols-4">
          <div className="col-span-2 md:col-span-1">
            <div className="flex items-center gap-2">
              <Logo className="size-10 rounded-sm" />
              <span className="font-display text-xl tracking-tight">KHAOPOD</span>
            </div>
            <p className="mt-3 max-w-xs text-sm text-muted-foreground">{t('footer.tagline')}</p>
          </div>

          <FooterCol
            title={t('footer.shop')}
            links={[
              { to: '/shop', label: t('footer.allCollection') },
              { to: '/shop', label: t('footer.new') },
              { to: '/shop', label: t('footer.bestSeller') },
            ]}
          />
          <FooterCol
            title={t('footer.help')}
            links={[
              { to: '/account', label: t('footer.myAccount') },
              { to: '/shop', label: t('footer.howToOrder') },
              { to: '/shop', label: t('footer.shipping') },
            ]}
          />
          <FooterCol
            title={t('footer.follow')}
            links={[
              { to: 'https://www.facebook.com/khaopodrises', label: 'Facebook' },
              { to: 'https://x.com/khaopoddddd', label: 'X (Twitter)' },
              { to: 'https://www.youtube.com/@khaopodtv', label: 'Youtube' },
            ]}
          />
        </div>

        <p className="mt-12 text-center text-xs text-muted-foreground">
          © {new Date().getFullYear()} Khaopod News Shop. {t('footer.rights')}
        </p>
      </div>
    </footer>
  )
}

function FooterCol({ title, links }: { title: string; links: { to: string; label: string }[] }) {
  return (
    <div>
      <p className="mb-3 text-xs font-bold uppercase tracking-widest">{title}</p>
      <ul className="flex flex-col gap-2">
        {links.map((l, i) => (
          <li key={i}>
            <FooterLink to={l.to}>{l.label}</FooterLink>
          </li>
        ))}
      </ul>
    </div>
  )
}

// ลิงก์ภายนอก (http...) เปิดแท็บใหม่ด้วย <a>; ลิงก์ภายในใช้ <Link> ของ router
function FooterLink({ to, children }: { to: string; children: ReactNode }) {
  const className = 'text-sm text-muted-foreground hover:text-foreground'
  if (to.startsWith('http')) {
    return (
      <a href={to} target="_blank" rel="noopener noreferrer" className={className}>
        {children}
      </a>
    )
  }
  return (
    <Link to={to} className={className}>
      {children}
    </Link>
  )
}
