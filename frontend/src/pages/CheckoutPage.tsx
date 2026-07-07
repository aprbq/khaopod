import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Spinner } from '@/components/ui/spinner'
import { AddressForm } from '@/features/addresses/AddressForm'
import { useAddresses, useCreateAddress, useSetDefaultAddress } from '@/features/addresses/hooks'
import { useCart } from '@/features/cart/hooks'
import { usePlaceOrder } from '@/features/orders/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { ApiError } from '@/lib/apiClient'
import { FLAT_SHIPPING_FEE } from '@/lib/constants'
import { formatBaht } from '@/lib/format'
import type { PaymentMethod } from '@/types/order'

export function CheckoutPage() {
  const { t } = useLang()
  useDocumentTitle(t('checkout.docTitle'))
  const navigate = useNavigate()

  const { data: cart, isLoading: cartLoading } = useCart()
  const { data: addresses, isLoading: addrLoading } = useAddresses()
  const createAddress = useCreateAddress()
  const setDefaultAddress = useSetDefaultAddress()
  const placeOrder = usePlaceOrder()

  const [addressId, setAddressId] = useState<number | null>(null)
  const [method, setMethod] = useState<PaymentMethod>('promptpay')
  const [note, setNote] = useState('')
  const [showAddressForm, setShowAddressForm] = useState(false)
  const [err, setErr] = useState('')

  // เลือกที่อยู่หลัก (หรืออันแรก) ให้อัตโนมัติเมื่อรายการโหลดเสร็จ
  useEffect(() => {
    if (!addresses || addressId !== null) return
    const preferred = addresses.find((a) => a.is_default) ?? addresses[0]
    if (preferred) setAddressId(preferred.id)
  }, [addresses, addressId])

  if (cartLoading || addrLoading) {
    return (
      <div className="grid min-h-[50vh] place-items-center">
        <Spinner className="size-6 text-muted-foreground" />
      </div>
    )
  }

  const items = cart?.items ?? []
  // สั่งซื้อสำเร็จจะ navigate ออกก่อนตะกร้าถูกล้าง — มาถึงหน้านี้ทั้งที่ตะกร้าว่าง = เข้าตรง ๆ
  if (items.length === 0 && !placeOrder.isPending && !placeOrder.isSuccess) {
    return (
      <div className="mx-auto grid min-h-[50vh] max-w-6xl place-items-center px-4 py-16 text-center">
        <div>
          <h1 className="font-display text-2xl uppercase">{t('checkout.title')}</h1>
          <p className="mt-2 text-sm text-muted-foreground">{t('checkout.emptyDesc')}</p>
          <Link to="/shop" className="mt-6 inline-block">
            <Button size="lg">{t('cart.browse')}</Button>
          </Link>
        </div>
      </div>
    )
  }

  const subtotal = cart?.subtotal ?? 0
  const total = Number(subtotal) + FLAT_SHIPPING_FEE

  const submit = () => {
    if (!addressId) {
      setErr(t('checkout.noAddress'))
      return
    }
    setErr('')
    placeOrder.mutate(
      { address_id: addressId, payment_method: method, customer_note: note || undefined },
      {
        onSuccess: (o) => navigate(`/orders/${o.order_number}`, { replace: true }),
        onError: (e) => setErr(e instanceof ApiError ? e.message : t('common.error')),
      },
    )
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      <h1 className="font-display text-2xl uppercase md:text-3xl">{t('checkout.title')}</h1>

      <div className="mt-6 grid gap-8 lg:grid-cols-[1fr_22rem]">
        <div className="flex flex-col gap-6">
          {/* ที่อยู่จัดส่ง */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">{t('checkout.shippingAddress')}</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-3">
              {(addresses ?? []).length === 0 && !showAddressForm && (
                <p className="text-sm text-muted-foreground">{t('checkout.noAddress')}</p>
              )}

              {(addresses ?? []).map((a) => (
                <label
                  key={a.id}
                  className={`flex cursor-pointer items-start gap-3 border p-3 text-sm transition-colors ${
                    addressId === a.id ? 'border-foreground' : 'border-border hover:border-muted-foreground'
                  }`}
                >
                  <input
                    type="radio"
                    name="address"
                    className="mt-1 accent-accent"
                    checked={addressId === a.id}
                    onChange={() => setAddressId(a.id)}
                  />
                  <span className="min-w-0 flex-1">
                    <span className="flex flex-wrap items-center gap-2 font-medium">
                      {a.recipient_name} · {a.phone}
                      {a.is_default && <Badge variant="outline">{t('checkout.defaultBadge')}</Badge>}
                    </span>
                    <span className="mt-0.5 block text-muted-foreground">
                      {a.address_line} {a.subdistrict} {a.district} {a.province} {a.postal_code}
                    </span>
                    {!a.is_default && (
                      <button
                        type="button"
                        className="mt-1.5 text-xs underline hover:text-accent disabled:opacity-40"
                        disabled={setDefaultAddress.isPending}
                        onClick={(e) => {
                          e.preventDefault() // อยู่ใน <label> — กันคลิกแล้วไปสลับ radio ด้วย
                          setErr('')
                          setDefaultAddress.mutate(a.id, {
                            onError: (er) =>
                              setErr(er instanceof ApiError ? er.message : t('common.error')),
                          })
                        }}
                      >
                        {t('addr.setDefault')}
                      </button>
                    )}
                  </span>
                </label>
              ))}

              {showAddressForm ? (
                <AddressForm
                  pending={createAddress.isPending}
                  onCancel={() => setShowAddressForm(false)}
                  onSubmit={(input) =>
                    createAddress.mutate(input, {
                      onSuccess: (a) => {
                        setAddressId(a.id)
                        setShowAddressForm(false)
                      },
                      onError: (e) => setErr(e instanceof ApiError ? e.message : t('common.error')),
                    })
                  }
                />
              ) : (
                <button
                  type="button"
                  className="self-start text-sm underline hover:text-accent"
                  onClick={() => setShowAddressForm(true)}
                >
                  + {t('checkout.addAddress')}
                </button>
              )}
            </CardContent>
          </Card>

          {/* วิธีชำระเงิน */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">{t('checkout.paymentMethod')}</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-2">
              {(
                [
                  ['promptpay', t('checkout.promptpay')],
                  ['bank_transfer', t('checkout.bankTransfer')],
                ] as [PaymentMethod, string][]
              ).map(([value, label]) => (
                <label key={value} className="flex cursor-pointer items-center gap-3 text-sm">
                  <input
                    type="radio"
                    name="payment_method"
                    className="accent-accent"
                    checked={method === value}
                    onChange={() => setMethod(value)}
                  />
                  {label}
                </label>
              ))}

              <label className="mt-3 flex flex-col gap-1.5 text-sm">
                {t('checkout.note')}
                <textarea
                  rows={2}
                  placeholder={t('checkout.notePh')}
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                  className="rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                />
              </label>
            </CardContent>
          </Card>
        </div>

        {/* สรุปคำสั่งซื้อ */}
        <aside className="h-fit border border-border p-5">
          <h2 className="font-display text-lg uppercase">{t('checkout.summary')}</h2>
          <ul className="mt-4 flex flex-col gap-2 text-sm">
            {items.map((it) => (
              <li key={it.id} className="flex justify-between gap-2">
                <span className="min-w-0 truncate text-muted-foreground">
                  {it.product_name} ({it.variant_name}
                  {it.color ? ` / ${it.color}` : ''}) × {it.quantity}
                </span>
                <span className="whitespace-nowrap">{formatBaht(it.line_total)}</span>
              </li>
            ))}
          </ul>
          <div className="mt-4 flex flex-col gap-1.5 border-t border-border pt-3 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t('cart.subtotal')}</span>
              <span>{formatBaht(subtotal)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t('checkout.shippingFee')}</span>
              <span>{formatBaht(FLAT_SHIPPING_FEE)}</span>
            </div>
            <div className="flex justify-between text-base font-bold">
              <span>{t('checkout.total')}</span>
              <span>{formatBaht(total)}</span>
            </div>
          </div>

          {err && (
            <p className="mt-4 rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {err}
            </p>
          )}

          <Button
            className="mt-5 w-full"
            size="lg"
            disabled={placeOrder.isPending || !addressId}
            onClick={submit}
          >
            {placeOrder.isPending && <Spinner />}
            {t('checkout.placeOrder')}
          </Button>
        </aside>
      </div>
    </div>
  )
}
