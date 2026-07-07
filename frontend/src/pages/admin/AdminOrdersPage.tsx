import { Link, useSearchParams } from 'react-router-dom'
import { Spinner } from '@/components/ui/spinner'
import { useAdminOrders } from '@/features/admin/hooks'
import { OrderStatusBadge } from '@/features/orders/OrderStatusBadge'
import { useLang } from '@/i18n/LanguageContext'
import type { TranslationKey } from '@/i18n/translations'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { formatBaht, formatDate } from '@/lib/format'
import { cn } from '@/lib/utils'
import type { OrderStatus } from '@/types/order'

const FILTERS: OrderStatus[] = ['pending', 'paid', 'preparing', 'shipped', 'delivered', 'cancelled']

export function AdminOrdersPage() {
  const { t, lang } = useLang()
  useDocumentTitle(t('admin.orders'))
  const [params, setParams] = useSearchParams()
  const status = (params.get('status') ?? '') as OrderStatus | ''
  const { data: orders, isLoading } = useAdminOrders(status || undefined)

  return (
    <div>
      <h1 className="font-display text-2xl uppercase md:text-3xl">{t('admin.orders')}</h1>

      {/* filter ตามสถานะ */}
      <div className="mt-4 flex flex-wrap gap-2">
        {['', ...FILTERS].map((s) => (
          <button
            key={s || 'all'}
            type="button"
            onClick={() => setParams(s ? { status: s } : {})}
            className={cn(
              'border px-3 py-1 text-xs font-semibold uppercase tracking-wide transition-colors',
              status === s
                ? 'border-foreground bg-foreground text-background'
                : 'border-border hover:border-foreground',
            )}
          >
            {s ? t(`orderStatus.${s}` as TranslationKey) : t('admin.filterAll')}
          </button>
        ))}
      </div>

      {isLoading ? (
        <div className="grid min-h-[30vh] place-items-center">
          <Spinner className="size-6 text-muted-foreground" />
        </div>
      ) : !orders || orders.length === 0 ? (
        <p className="mt-8 text-sm text-muted-foreground">{t('admin.empty')}</p>
      ) : (
        <div className="mt-4 overflow-x-auto border border-border bg-background">
          <table className="w-full min-w-[42rem] text-left text-sm">
            <thead className="border-b border-border text-xs uppercase tracking-wide text-muted-foreground">
              <tr>
                <th className="px-4 py-3">{t('admin.colOrder')}</th>
                <th className="px-4 py-3">{t('admin.colCustomer')}</th>
                <th className="px-4 py-3">{t('admin.colDate')}</th>
                <th className="px-4 py-3 text-right">{t('admin.colTotal')}</th>
                <th className="px-4 py-3">{t('admin.colStatus')}</th>
                <th className="px-4 py-3">{t('admin.colPayment')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {orders.map((o) => (
                <tr key={o.order_number} className="transition-colors hover:bg-muted/50">
                  <td className="px-4 py-3 font-mono font-semibold">
                    <Link to={`/admin/orders/${o.order_number}`} className="hover:underline">
                      {o.order_number}
                    </Link>
                  </td>
                  <td className="max-w-[14rem] truncate px-4 py-3 text-muted-foreground">
                    {o.user_email}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-muted-foreground">
                    {formatDate(o.placed_at, lang)}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-right font-semibold">
                    {formatBaht(o.total_amount)}
                  </td>
                  <td className="px-4 py-3">
                    <OrderStatusBadge status={o.status} />
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {t(`payStatus.${o.payment_status}` as TranslationKey)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
