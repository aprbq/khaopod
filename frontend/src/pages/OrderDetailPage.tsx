import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { ConfirmDialog } from '@/components/ConfirmDialog'
import { useCancelOrder, useOrder, useSubmitPayment } from '@/features/orders/hooks'
import { useLang } from '@/i18n/LanguageContext'
import type { TranslationKey } from '@/i18n/translations'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { ApiError } from '@/lib/apiClient'
import { SHOP_PAYMENT } from '@/lib/constants'
import { formatBaht, formatDate } from '@/lib/format'
import { OrderStatusBadge } from '@/features/orders/OrderStatusBadge'
import type { Order, PaymentMethod } from '@/types/order'

const MAX_SLIP_BYTES = 5 * 1024 * 1024 // ต้องตรงกับเพดานฝั่ง backend (docs §10.1)

export function OrderDetailPage() {
  const { t, lang } = useLang()
  const { orderNumber = '' } = useParams()
  useDocumentTitle(orderNumber)
  const { data: order, isLoading } = useOrder(orderNumber)
  const cancelOrder = useCancelOrder()
  const [confirmCancel, setConfirmCancel] = useState(false)
  const [err, setErr] = useState('')

  if (isLoading) {
    return (
      <div className="grid min-h-[50vh] place-items-center">
        <Spinner className="size-6 text-muted-foreground" />
      </div>
    )
  }
  if (!order) {
    return (
      <div className="mx-auto grid min-h-[50vh] max-w-6xl place-items-center px-4 py-16 text-center">
        <div>
          <h1 className="font-display text-2xl uppercase">{t('order.notFound')}</h1>
          <Link to="/orders" className="mt-4 inline-block text-sm underline hover:text-accent">
            {t('order.backToOrders')}
          </Link>
        </div>
      </div>
    )
  }

  const canCancel = order.status === 'pending' || order.status === 'paid'
  const ship = order.shipping_address

  return (
    <div className="mx-auto max-w-3xl px-4 py-8">
      <Link to="/orders" className="text-sm text-muted-foreground underline hover:text-foreground">
        {t('order.backToOrders')}
      </Link>

      <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="font-mono text-xl font-bold md:text-2xl">{order.order_number}</h1>
          <p className="mt-1 text-xs text-muted-foreground">
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

      <div className="mt-6 flex flex-col gap-6">
        {/* รายการสินค้า + ยอดรวม */}
        <Card>
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
          </CardContent>
        </Card>

        {/* ที่อยู่จัดส่ง (snapshot) */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">{t('order.shipTo')}</CardTitle>
          </CardHeader>
          <CardContent className="text-sm">
            <p className="font-medium">
              {ship.recipient} · {ship.phone}
            </p>
            <p className="mt-1 text-muted-foreground">
              {ship.address} {ship.subdistrict} {ship.district} {ship.province} {ship.postal_code}
            </p>
          </CardContent>
        </Card>

        {/* การชำระเงิน */}
        <PaymentSection order={order} onError={setErr} />

        {canCancel && (
          <button
            type="button"
            className="self-start text-sm text-muted-foreground underline hover:text-destructive disabled:opacity-40"
            disabled={cancelOrder.isPending}
            onClick={() => setConfirmCancel(true)}
          >
            {t('order.cancel')}
          </button>
        )}
      </div>

      {confirmCancel && (
        <ConfirmDialog
          title={t('order.cancelConfirmTitle')}
          desc={t('order.cancelConfirmDesc')}
          confirmLabel={t('order.cancel')}
          cancelLabel={t('common.cancel')}
          onCancel={() => setConfirmCancel(false)}
          onConfirm={() => {
            setConfirmCancel(false)
            setErr('')
            cancelOrder.mutate(
              { orderNumber: order.order_number },
              { onError: (e) => setErr(e instanceof ApiError ? e.message : t('common.error')) },
            )
          }}
        />
      )}
    </div>
  )
}

// ส่วนชำระเงิน — โชว์ช่องทางโอน + ฟอร์มแนบสลิปเมื่อยังไม่จ่าย, สถานะเมื่อแจ้งแล้ว
function PaymentSection({ order, onError }: { order: Order; onError: (msg: string) => void }) {
  const { t } = useLang()
  const submitPayment = useSubmitPayment()
  const [method, setMethod] = useState<PaymentMethod>(order.payment_method ?? 'promptpay')
  const [slip, setSlip] = useState<File | null>(null)
  const [ref, setRef] = useState('')
  const [formErr, setFormErr] = useState('')

  const status = order.payment_status

  if (status === 'paid') {
    return (
      <p className="rounded-md bg-accent/10 px-3 py-2 text-sm text-accent">{t('order.paidDone')}</p>
    )
  }
  if (status === 'pending_review') {
    return (
      <p className="rounded-md bg-accent/10 px-3 py-2 text-sm text-accent">
        {t('order.pendingReview')}
        {order.payment?.slip_url && (
          <a
            href={order.payment.slip_url}
            target="_blank"
            rel="noreferrer"
            className="ml-2 underline"
          >
            {t('order.viewSlip')}
          </a>
        )}
      </p>
    )
  }
  // ยกเลิก/คืนเงินแล้ว ไม่ต้องโชว์ฟอร์มจ่าย
  if (order.status !== 'pending') {
    return null
  }

  const submit = () => {
    if (!slip) {
      setFormErr(t('order.slipRequired'))
      return
    }
    if (slip.size > MAX_SLIP_BYTES) {
      setFormErr(t('order.slipTooBig'))
      return
    }
    setFormErr('')
    onError('')
    submitPayment.mutate(
      {
        orderNumber: order.order_number,
        input: { method, amount: order.total_amount, transaction_ref: ref || undefined, slip },
      },
      { onError: (e) => onError(e instanceof ApiError ? e.message : t('common.error')) },
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">{t('order.paymentTitle')}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4 text-sm">
        {status === 'failed' && (
          <p className="rounded-md bg-destructive/10 px-3 py-2 text-destructive">
            {t('order.paymentFailed')}
          </p>
        )}

        <div className="flex flex-col gap-2">
          {(
            [
              ['promptpay', t('checkout.promptpay')],
              ['bank_transfer', t('checkout.bankTransfer')],
            ] as [PaymentMethod, string][]
          ).map(([value, label]) => (
            <label key={value} className="flex cursor-pointer items-center gap-3">
              <input
                type="radio"
                name="pay_method"
                className="accent-accent"
                checked={method === value}
                onChange={() => setMethod(value)}
              />
              {label}
            </label>
          ))}
        </div>

        {/* ช่องทางโอนของร้าน */}
        <div className="rounded-md bg-muted px-4 py-3">
          {method === 'promptpay' ? (
            <>
              <p className="font-medium">{t('order.payViaPromptpay')}</p>
              <p className="mt-1">{t('order.promptpayId', { id: SHOP_PAYMENT.promptpayId })}</p>
            </>
          ) : (
            <>
              <p className="font-medium">{t('order.payViaBank')}</p>
              <p className="mt-1">
                {t('order.bankAccount', {
                  bank: SHOP_PAYMENT.bankName,
                  no: SHOP_PAYMENT.bankAccountNo,
                  name: SHOP_PAYMENT.bankAccountName,
                })}
              </p>
            </>
          )}
          <p className="mt-2">
            {t('order.amountToPay')}:{' '}
            <span className="text-base font-bold">{formatBaht(order.total_amount)}</span>
          </p>
        </div>

        <div className="flex flex-col gap-1.5">
          <Label htmlFor="slip">{t('order.slip')}</Label>
          <input
            id="slip"
            type="file"
            accept="image/jpeg,image/png,image/webp"
            className="text-sm file:mr-3 file:cursor-pointer file:rounded-md file:border-0 file:bg-primary file:px-3 file:py-1.5 file:text-sm file:font-semibold file:text-primary-foreground"
            onChange={(e) => setSlip(e.target.files?.[0] ?? null)}
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <Label htmlFor="txref">{t('order.transactionRef')}</Label>
          <Input id="txref" value={ref} onChange={(e) => setRef(e.target.value)} />
        </div>

        {formErr && <p className="text-destructive">{formErr}</p>}

        <Button className="self-start" disabled={submitPayment.isPending} onClick={submit}>
          {submitPayment.isPending && <Spinner />}
          {t('order.submitPayment')}
        </Button>
      </CardContent>
    </Card>
  )
}
