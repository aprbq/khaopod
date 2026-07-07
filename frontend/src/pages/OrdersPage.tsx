import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { OrderStatusBadge } from '@/features/orders/OrderStatusBadge'
import { useOrders } from '@/features/orders/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { formatBaht, formatDate } from '@/lib/format'

export function OrdersPage() {
  const { t, lang } = useLang()
  useDocumentTitle(t('orders.docTitle'))
  const { data: orders, isLoading } = useOrders()

  if (isLoading) {
    return (
      <div className="grid min-h-[50vh] place-items-center">
        <Spinner className="size-6 text-muted-foreground" />
      </div>
    )
  }

  if (!orders || orders.length === 0) {
    return (
      <div className="mx-auto grid min-h-[50vh] max-w-6xl place-items-center px-4 py-16 text-center">
        <div>
          <h1 className="font-display text-2xl uppercase">{t('orders.title')}</h1>
          <p className="mt-2 text-sm text-muted-foreground">{t('orders.empty')}</p>
          <Link to="/shop" className="mt-6 inline-block">
            <Button size="lg">{t('orders.browse')}</Button>
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-3xl px-4 py-8">
      <h1 className="font-display text-2xl uppercase md:text-3xl">{t('orders.title')}</h1>

      <ul className="mt-6 divide-y divide-border border-y border-border">
        {orders.map((o) => (
          <li key={o.order_number}>
            <Link
              to={`/orders/${o.order_number}`}
              className="flex flex-wrap items-center justify-between gap-3 py-4 transition-colors hover:bg-muted/50"
            >
              <div className="min-w-0">
                <p className="font-mono text-sm font-semibold">{o.order_number}</p>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  {t('order.placedAt', { date: formatDate(o.placed_at, lang) })} ·{' '}
                  {t('common.items', { n: o.items.reduce((n, it) => n + it.quantity, 0) })}
                </p>
              </div>
              <div className="flex items-center gap-3">
                <OrderStatusBadge status={o.status} />
                <span className="font-semibold">{formatBaht(o.total_amount)}</span>
              </div>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  )
}
