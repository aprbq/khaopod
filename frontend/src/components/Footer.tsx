import { Link } from 'react-router-dom'

export function Footer() {
  return (
    <footer className="border-t border-border">
      <div className="mx-auto max-w-6xl px-4 py-12">
        <div className="grid grid-cols-2 gap-8 md:grid-cols-4">
          <div className="col-span-2 md:col-span-1">
            <p className="font-display text-xl tracking-tight">
              KHAOPOD<span className="text-accent">.</span>
            </p>
            <p className="mt-2 max-w-xs text-sm text-muted-foreground">
              ร้านค้าอย่างเป็นทางการของเพจกองบัญชาการข่าวปด
            </p>
          </div>

          <FooterCol
            title="ช้อป"
            links={[
              { to: '/shop', label: 'สินค้าทั้งหมด' },
              { to: '/shop', label: 'มาใหม่' },
              { to: '/shop', label: 'ขายดี' },
            ]}
          />
          <FooterCol
            title="ช่วยเหลือ"
            links={[
              { to: '/account', label: 'บัญชีของฉัน' },
              { to: '/shop', label: 'วิธีสั่งซื้อ' },
              { to: '/shop', label: 'การจัดส่ง' },
            ]}
          />
          <FooterCol
            title="ติดตาม"
            links={[
              { to: '/', label: 'Facebook' },
              { to: '/', label: 'Instagram' },
              { to: '/', label: 'TikTok' },
            ]}
          />
        </div>

        <p className="mt-12 text-xs text-muted-foreground">
          © {new Date().getFullYear()} Khaopod News Shop. สงวนลิขสิทธิ์.
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
            <Link to={l.to} className="text-sm text-muted-foreground hover:text-foreground">
              {l.label}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  )
}
