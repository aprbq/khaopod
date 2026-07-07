import { Link } from 'react-router-dom'
import { Spinner } from '@/components/ui/spinner'
import { useAdminSummary } from '@/features/admin/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { formatBaht } from '@/lib/format'

export function AdminDashboardPage() {
  const { t } = useLang()
  useDocumentTitle(t('admin.dashboard'))
  const { data, isLoading } = useAdminSummary()

  if (isLoading || !data) {
    return (
      <div className="grid min-h-[40vh] place-items-center">
        <Spinner className="size-6 text-muted-foreground" />
      </div>
    )
  }

  const stats = [
    { label: t('admin.summaryOrdersTotal'), value: String(data.orders_total), to: '/admin/orders' },
    {
      label: t('admin.summaryPending'),
      value: String(data.orders_pending),
      to: '/admin/orders?status=pending',
    },
    {
      label: t('admin.summarySlips'),
      value: String(data.payments_pending_review),
      to: '/admin/orders?status=pending',
      highlight: data.payments_pending_review > 0,
    },
    { label: t('admin.summaryRevenue'), value: formatBaht(data.revenue_paid) },
  ]

  return (
    <div>
      <h1 className="font-display text-2xl uppercase md:text-3xl">{t('admin.dashboard')}</h1>

      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {stats.map((s) => {
          const card = (
            <div
              className={`h-full border bg-background p-5 ${
                s.highlight ? 'border-accent' : 'border-border'
              }`}
            >
              <p className="text-xs uppercase tracking-wide text-muted-foreground">{s.label}</p>
              <p className="mt-2 text-2xl font-bold">{s.value}</p>
            </div>
          )
          return s.to ? (
            <Link key={s.label} to={s.to} className="transition-opacity hover:opacity-80">
              {card}
            </Link>
          ) : (
            <div key={s.label}>{card}</div>
          )
        })}
      </div>
    </div>
  )
}
