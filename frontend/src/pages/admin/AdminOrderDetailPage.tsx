import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Spinner } from '@/components/ui/spinner'
import { ConfirmDialog } from '@/components/ConfirmDialog'
import { useAdminOrder, useAdminUpdateStatus, useAdminVerifyPayment } from '@/features/admin/hooks'
import { OrderStatusBadge } from '@/features/orders/OrderStatusBadge'
import { useLang } from '@/i18n/LanguageContext'
import type { TranslationKey } from '@/i18n/translations'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { ApiError } from '@/lib/apiClient'
import { formatBaht, formatDate } from '@/lib/format'
import type { OrderStatus } from '@/types/order'

// สถานะที่แอดมินตั้งได้ (ตรงกับ ValidAdminStatus ฝั่ง backend)
const ADMIN_STATUSES: OrderStatus[] = [
  'paid',
  'preparing',
  'shipped',
  'delivered',
  'completed',
  'cancelled',
  'refunded',
]

export function AdminOrderDetailPage() {
  const { t, lang } = useLang()
  const { orderNumber = '' } = useParams()
  useDocumentTitle(orderNumber)
  const { data: order, isLoading } = useAdminOrder(orderNumber)
  const verifyPayment = useAdminVerifyPayment()
  const updateStatus = useAdminUpdateStatus()

  const [verdict, setVerdict] = useState<'paid' | 'failed' | null>(null)
  const [newStatus, setNewStatus] = useState<OrderStatus | ''>('')
  const [note, setNote] = useState('')
  const [err, setErr] = useState('')

  if (isLoading) {
    return (
      <div className="grid min-h-[40vh] place-items-center">
        <Spinner className="size-6 text-muted-foreground" />
      </div>
    )
  }
  if (!order) {
    return (
      <div>
        <p className="text-sm text-muted-foreground">{t('order.notFound')}</p>
        <Link to="/admin/orders" className="mt-2 inline-block text-sm underline">
          {t('order.backToOrders')}
        </Link>
      </div>
    )
  }

  const onApiError = (e: unknown) => setErr(e instanceof ApiError ? e.message : t('common.error'))
  const payment = order.payment
  const ship = order.shipping_address

  return (
    <div>
      <Link
        to="/admin/orders"
        className="text-sm text-muted-foreground underline hover:text-foreground"
      >
        {t('order.backToOrders')}
      </Link>

      <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="font-mono text-xl font-bold md:text-2xl">{order.order_number}</h1>
          <p className="mt-1 text-xs text-muted-foreground">
            {t('admin.colCustomer')}: {order.user_email} ·{' '}
            {t('order.placedAt', { date: formatDate(order.placed_at, lang) })}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <OrderStatusBadge status={order.status} />
          <Badge variant="outline">
            {t(`payStatus.${order.payment_status}` as TranslationKey)}
          </Badge>
        </div>
      </div>

      {err && (
        <p className="mt-4 rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{err}</p>
      )}

      <div className="mt-6 grid gap-6 lg:grid-cols-2">
        {/* รายการ + ยอด */}
        <Card className="bg-background">
          <CardContent className="pt-6">
            <ul className="flex flex-col gap-2 text-sm">
              {order.items.map((it, i) => (
                <li key={i} className="flex justify-between gap-2">
                  <span className="min-w-0 text-muted-foreground">
                    {it.product_name} ({it.variant_name}) × {it.quantity}
                  </span>
                  <span className="whitespace-nowrap">{formatBaht(it.line_total)}</span>
                </li>
              ))}
            </ul>
            <div className="mt-4 flex flex-col gap-1.5 border-t border-border pt-3 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">{t('cart.subtotal')}</span>
                <span>{formatBaht(order.subtotal)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">{t('checkout.shippingFee')}</span>
                <span>{formatBaht(order.shipping_fee)}</span>
              </div>
              <div className="flex justify-between text-base font-bold">
                <span>{t('checkout.total')}</span>
                <span>{formatBaht(order.total_amount)}</span>
              </div>
            </div>
            <div className="mt-4 border-t border-border pt-3 text-sm">
              <p className="font-medium">
                {ship.recipient} · {ship.phone}
              </p>
              <p className="mt-1 text-muted-foreground">
                {ship.address} {ship.subdistrict} {ship.district} {ship.province} {ship.postal_code}
              </p>
              {order.customer_note && (
                <p className="mt-2 text-muted-foreground">📝 {order.customer_note}</p>
              )}
            </div>
          </CardContent>
        </Card>

        <div className="flex flex-col gap-6">
          {/* ตรวจสลิป */}
          <Card className="bg-background">
            <CardHeader>
              <CardTitle className="text-lg">{t('admin.slipReview')}</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-3 text-sm">
              {!payment ? (
                <p className="text-muted-foreground">{t('admin.noPayment')}</p>
              ) : (
                <>
                  <p className="text-muted-foreground">
                    {t(`payStatus.${payment.status}` as TranslationKey)} ·{' '}
                    {payment.method === 'promptpay'
                      ? t('checkout.promptpay')
                      : t('checkout.bankTransfer')}{' '}
                    · {formatBaht(payment.amount)}
                    {payment.transaction_ref && ` · ${payment.transaction_ref}`}
                  </p>
                  {payment.slip_url && (
                    <a href={payment.slip_url} target="_blank" rel="noreferrer">
                      <img
                        src={payment.slip_url}
                        alt={t('order.slip')}
                        loading="lazy"
                        className="max-h-80 border border-border object-contain"
                      />
                    </a>
                  )}
                  {payment.status === 'pending_review' && (
                    <div className="flex gap-3">
                      <Button
                        disabled={verifyPayment.isPending}
                        onClick={() => setVerdict('paid')}
                      >
                        {t('admin.approve')}
                      </Button>
                      <Button
                        variant="destructive"
                        disabled={verifyPayment.isPending}
                        onClick={() => setVerdict('failed')}
                      >
                        {t('admin.reject')}
                      </Button>
                    </div>
                  )}
                </>
              )}
            </CardContent>
          </Card>

          {/* อัปเดตสถานะ */}
          <Card className="bg-background">
            <CardHeader>
              <CardTitle className="text-lg">{t('admin.updateStatus')}</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-3 text-sm">
              <select
                value={newStatus}
                onChange={(e) => setNewStatus(e.target.value as OrderStatus)}
                className="h-10 rounded-md border border-input bg-background px-3 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              >
                <option value="" disabled>
                  — {t('admin.colStatus')} —
                </option>
                {ADMIN_STATUSES.map((s) => (
                  <option key={s} value={s}>
                    {t(`orderStatus.${s}` as TranslationKey)}
                  </option>
                ))}
              </select>
              <input
                type="text"
                placeholder={t('admin.statusNote')}
                value={note}
                onChange={(e) => setNote(e.target.value)}
                className="h-10 rounded-md border border-input bg-background px-3 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              />
              <Button
                className="self-start"
                disabled={!newStatus || updateStatus.isPending}
                onClick={() => {
                  if (!newStatus) return
                  setErr('')
                  updateStatus.mutate(
                    { orderNumber: order.order_number, status: newStatus, note },
                    { onSuccess: () => setNewStatus(''), onError: onApiError },
                  )
                }}
              >
                {updateStatus.isPending && <Spinner />}
                {t('profile.save')}
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>

      {verdict && payment && (
        <ConfirmDialog
          title={verdict === 'paid' ? t('admin.approveConfirmTitle') : t('admin.rejectConfirmTitle')}
          desc={verdict === 'paid' ? t('admin.approveConfirmDesc') : t('admin.rejectConfirmDesc')}
          confirmLabel={verdict === 'paid' ? t('admin.approve') : t('admin.reject')}
          confirmVariant={verdict === 'paid' ? 'default' : 'destructive'}
          cancelLabel={t('common.cancel')}
          onCancel={() => setVerdict(null)}
          onConfirm={() => {
            const status = verdict
            setVerdict(null)
            setErr('')
            verifyPayment.mutate({ paymentId: payment.id, status }, { onError: onApiError })
          }}
        />
      )}
    </div>
  )
}
