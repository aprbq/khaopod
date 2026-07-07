import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { ConfirmDialog } from '@/components/ConfirmDialog'
import { IconBag } from '@/components/icons'
import { useAuth } from '@/features/auth/AuthContext'
import { useCart, useClearCart, useRemoveCartItem, useUpdateCartItem } from '@/features/cart/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { ApiError } from '@/lib/apiClient'
import { formatBaht } from '@/lib/format'

// เชลล์กลางจอ ใช้กับสถานะ empty / ต้องล็อกอิน
function CartMessage({
  title,
  desc,
  cta,
}: {
  title: string
  desc: string
  cta: { to: string; label: string }
}) {
  return (
    <div className="mx-auto grid min-h-[50vh] max-w-6xl place-items-center px-4 py-16 text-center">
      <div>
        <IconBag className="mx-auto size-10 text-muted-foreground" />
        <h1 className="mt-4 font-display text-2xl uppercase">{title}</h1>
        <p className="mt-2 text-sm text-muted-foreground">{desc}</p>
        <Link to={cta.to} className="mt-6 inline-block">
          <Button size="lg">{cta.label}</Button>
        </Link>
      </div>
    </div>
  )
}

export function CartPage() {
  const { t } = useLang()
  useDocumentTitle(t('cart.docTitle'))
  const { status } = useAuth()
  const { data: cart, isLoading } = useCart()
  const updateItem = useUpdateCartItem()
  const removeItem = useRemoveCartItem()
  const clearCart = useClearCart()
  const [err, setErr] = useState('')
  const [info, setInfo] = useState('')
  // เป้าหมายที่รอผู้ใช้กดยืนยันก่อนลบ — null = ไม่มี dialog เปิดอยู่
  const [confirm, setConfirm] = useState<
    { kind: 'item'; itemId: number; name: string } | { kind: 'clear' } | null
  >(null)

  const handleConfirm = () => {
    if (!confirm) return
    setErr('')
    if (confirm.kind === 'item') removeItem.mutate({ itemId: confirm.itemId })
    else clearCart.mutate()
    setConfirm(null)
  }

  const setQty = (itemId: number, quantity: number) => {
    setErr('')
    updateItem.mutate(
      { itemId, quantity },
      { onError: (e) => setErr(e instanceof ApiError ? e.message : t('cart.outOfStock')) },
    )
  }

  if (status === 'unauthenticated') {
    return (
      <CartMessage
        title={t('cart.title')}
        desc={t('cart.loginRequired')}
        cta={{ to: '/login', label: t('cart.login') }}
      />
    )
  }
  if (status === 'loading' || isLoading) {
    return (
      <div className="grid min-h-[50vh] place-items-center">
        <Spinner className="size-6 text-muted-foreground" />
      </div>
    )
  }

  const items = cart?.items ?? []
  if (items.length === 0) {
    return (
      <CartMessage
        title={t('cart.emptyTitle')}
        desc={t('cart.emptyDesc')}
        cta={{ to: '/shop', label: t('cart.browse') }}
      />
    )
  }

  const busy = updateItem.isPending || removeItem.isPending

  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      <h1 className="font-display text-2xl uppercase md:text-3xl">{t('cart.title')}</h1>

      {err && (
        <p className="mt-4 rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{err}</p>
      )}
      {info && <p className="mt-4 rounded-md bg-accent/10 px-3 py-2 text-sm text-accent">{info}</p>}

      <div className="mt-6 grid gap-8 lg:grid-cols-[1fr_20rem]">
        {/* รายการสินค้า */}
        <div>
          <ul className="divide-y divide-border border-y border-border">
            {items.map((it) => (
              <li key={it.id} className="flex gap-4 py-4">
                <div className="size-20 shrink-0 overflow-hidden bg-muted">
                  {it.image && (
                    <img
                      src={it.image}
                      alt={it.product_name}
                      loading="lazy"
                      className="h-full w-full object-cover"
                    />
                  )}
                </div>

                <div className="min-w-0 flex-1">
                  <div className="flex items-start justify-between gap-2">
                    <p className="font-medium">{it.product_name}</p>
                    <p className="whitespace-nowrap font-semibold">{formatBaht(it.line_total)}</p>
                  </div>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {it.variant_name}
                    {it.color ? ` · ${it.color}` : ''} · {formatBaht(it.unit_price)}
                  </p>
                  {!it.in_stock && (
                    <p className="mt-1 text-xs font-semibold text-destructive">
                      {t('cart.outOfStock')}
                    </p>
                  )}

                  <div className="mt-3 flex items-center gap-4">
                    <div className="flex items-center border border-input">
                      <button
                        type="button"
                        aria-label="-"
                        className="px-3 py-1 text-base leading-none disabled:opacity-40"
                        disabled={it.quantity <= 1 || busy}
                        onClick={() => setQty(it.id, it.quantity - 1)}
                      >
                        −
                      </button>
                      <span className="min-w-8 text-center text-sm">{it.quantity}</span>
                      <button
                        type="button"
                        aria-label="+"
                        className="px-3 py-1 text-base leading-none disabled:opacity-40"
                        disabled={busy}
                        onClick={() => setQty(it.id, it.quantity + 1)}
                      >
                        +
                      </button>
                    </div>
                    <button
                      type="button"
                      className="text-sm text-muted-foreground underline hover:text-destructive disabled:opacity-40"
                      disabled={busy}
                      onClick={() =>
                        setConfirm({ kind: 'item', itemId: it.id, name: it.product_name })
                      }
                    >
                      {t('cart.remove')}
                    </button>
                  </div>
                </div>
              </li>
            ))}
          </ul>

          <button
            type="button"
            className="mt-4 text-sm text-muted-foreground underline hover:text-foreground disabled:opacity-40"
            disabled={clearCart.isPending}
            onClick={() => setConfirm({ kind: 'clear' })}
          >
            {t('cart.clear')}
          </button>
        </div>

        {/* สรุปยอด */}
        <aside className="h-fit border border-border p-5">
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">{t('cart.subtotal')}</span>
            <span className="text-lg font-bold">{formatBaht(cart?.subtotal ?? 0)}</span>
          </div>
          {/* ระบบชำระเงินยังไม่เปิด — เฟสถัดไป */}
          <Button className="mt-5 w-full" size="lg" onClick={() => setInfo(t('cart.checkoutSoon'))}>
            {t('cart.checkout')}
          </Button>
        </aside>
      </div>

      {confirm && (
        <ConfirmDialog
          title={confirm.kind === 'item' ? t('cart.removeConfirmTitle') : t('cart.clearConfirmTitle')}
          desc={
            confirm.kind === 'item'
              ? t('cart.removeConfirmDesc', { name: confirm.name })
              : t('cart.clearConfirmDesc')
          }
          confirmLabel={confirm.kind === 'item' ? t('cart.remove') : t('cart.clear')}
          cancelLabel={t('common.cancel')}
          onConfirm={handleConfirm}
          onCancel={() => setConfirm(null)}
        />
      )}
    </div>
  )
}
