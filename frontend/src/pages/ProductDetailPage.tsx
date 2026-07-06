import { useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ImageLightbox } from '@/components/ImageLightbox'
import { Spinner } from '@/components/ui/spinner'
import { useAuth } from '@/features/auth/AuthContext'
import { useAddToCart } from '@/features/cart/hooks'
import { useProduct } from '@/features/products/hooks'
import { ApiError } from '@/lib/apiClient'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { formatBaht } from '@/lib/format'
import { cn } from '@/lib/utils'
import type { ProductVariant } from '@/types/product'

// สีของ swatch ตามชื่อสี (ค่าจาก backend ไม่แปล) — ชื่อที่ไม่รู้จักใช้เทากลาง ๆ
const SWATCH_HEX: Record<string, string> = {
  ขาว: '#ffffff',
  ดำ: '#1a1a1a',
  แดง: '#dc2626',
  น้ำเงิน: '#1d4ed8',
  เขียว: '#16a34a',
  เหลือง: '#eab308',
  เทา: '#9ca3af',
  ครีม: '#f5f0e6',
}
const swatchColor = (name: string) => SWATCH_HEX[name] ?? '#d4d4d8'

export function ProductDetailPage() {
  const { t } = useLang()
  const { slug } = useParams<{ slug: string }>()
  const { data, isLoading, isError } = useProduct(slug ?? '')
  const [size, setSize] = useState<string | null>(null)
  const [color, setColor] = useState<string | null>(null)
  const [activeIdx, setActiveIdx] = useState(0)
  const [zoomOpen, setZoomOpen] = useState(false)
  const [notice, setNotice] = useState('')
  const navigate = useNavigate()
  const { status } = useAuth()
  const addToCartMut = useAddToCart()
  useDocumentTitle(data?.name)

  if (isLoading) {
    return (
      <div className="grid min-h-[50vh] place-items-center">
        <Spinner className="size-6 text-muted-foreground" />
      </div>
    )
  }
  if (isError || !data) {
    return (
      <div className="mx-auto max-w-6xl px-4 py-24 text-center">
        <p className="text-sm text-muted-foreground">{t('product.notFound')}</p>
        <Link to="/shop" className="mt-4 inline-block text-sm font-semibold underline">
          {t('product.backCollection')}
        </Link>
      </div>
    )
  }

  const addToCart = () => {
    if (status !== 'authenticated') {
      navigate('/login') // ต้องล็อกอินก่อนถึงจะมีตะกร้า
      return
    }
    if (!selected) return
    addToCartMut.mutate(
      { variantId: selected.id, quantity: 1 },
      {
        onSuccess: () => setNotice(t('product.added')),
        onError: (e) => setNotice(e instanceof ApiError ? e.message : t('product.outOfStock')),
      },
    )
  }

  const images = data.images ?? []
  const activeImage = images[activeIdx]

  // แยกไซซ์กับสีออกเป็นคนละมิติจากรายการ variant
  const variants = data.variants
  const sizes = [...new Set(variants.map((v) => v.variant_name))]
  const colors = [...new Set(variants.map((v) => v.color).filter((c): c is string => !!c))]
  const hasColors = colors.length > 0

  // variant ที่ตรงกับไซซ์+สีที่เลือก (ถ้าสินค้าไม่มีสี ก็แค่ตรงไซซ์)
  const selected: ProductVariant | null =
    variants.find((v) => v.variant_name === size && (!hasColors || v.color === color)) ?? null

  // ไซซ์นี้มีของไหม (ถ้าเลือกสีไว้แล้ว ดูเฉพาะสีนั้น)
  const sizeInStock = (s: string) =>
    variants.some((v) => v.variant_name === s && v.stock_quantity > 0 && (!hasColors || !color || v.color === color))
  // สีนี้มีของไหม (ถ้าเลือกไซซ์ไว้แล้ว ดูเฉพาะไซซ์นั้น)
  const colorInStock = (c: string) =>
    variants.some((v) => v.color === c && v.stock_quantity > 0 && (!size || v.variant_name === size))

  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      <Link to="/shop" className="text-xs font-semibold uppercase tracking-widest text-muted-foreground hover:text-foreground">
        {t('product.backCollectionShort')}
      </Link>

      {/* min-w-0 บน grid item: กัน CSS grid ขยายคอลัมน์ตามความกว้างจริงของรูป (intrinsic) จนล้นจอ */}
      <div className="mt-6 grid gap-8 md:grid-cols-2">
        {/* รูปสินค้า — แกลเลอรีเลื่อนดูได้ทั้งหมด */}
        <div className="min-w-0">
          <div className="relative aspect-square overflow-hidden bg-muted">
            {activeImage ? (
              // กดเพื่อเปิดดูรูปเต็มจอ + ซูม
              <button
                type="button"
                onClick={() => setZoomOpen(true)}
                aria-label={t('product.viewLarge')}
                className="block h-full w-full cursor-zoom-in"
              >
                <img src={activeImage.url} alt={data.name} className="h-full w-full object-cover" />
              </button>
            ) : (
              // ไม่มีรูป → placeholder แบบ typographic
              <div className="flex h-full w-full items-center justify-center p-8">
                <span className="text-center text-base font-semibold uppercase tracking-widest text-muted-foreground/50">
                  {data.name}
                </span>
              </div>
            )}
            {!data.in_stock && (
              <div className="absolute left-0 top-0 p-4">
                <Badge>{t('product.outOfStock')}</Badge>
              </div>
            )}
          </div>

          {/* แถบรูปย่อ — เลื่อนแนวนอนดูได้ทุกรูป กดเพื่อเปลี่ยนรูปหลัก */}
          {images.length > 1 && (
            <div className="mt-3 flex gap-2 overflow-x-auto pb-1">
              {images.map((img, i) => (
                <button
                  key={img.id}
                  type="button"
                  onClick={() => setActiveIdx(i)}
                  aria-label={t('product.viewImageN', { n: i + 1 })}
                  aria-current={i === activeIdx}
                  className={cn(
                    'aspect-square w-16 shrink-0 overflow-hidden border transition-colors',
                    i === activeIdx ? 'border-primary' : 'border-transparent hover:border-input',
                  )}
                >
                  <img src={img.url} alt="" loading="lazy" className="h-full w-full object-cover" />
                </button>
              ))}
            </div>
          )}
        </div>

        {/* รายละเอียด */}
        <div>
          <p className="text-xs uppercase tracking-widest text-muted-foreground">{data.category?.name}</p>
          <h1 className="mt-2 font-display text-2xl uppercase md:text-3xl">{data.name}</h1>
          <p className="mt-4 text-xl font-bold">
            {formatBaht(selected ? selected.price : data.price_range.min)}
          </p>

          {/* เลือกสี (เฉพาะสินค้าที่มีตัวเลือกสี) — swatch แยกจากไซซ์ */}
          {hasColors && (
            <div className="mt-6">
              <p className="mb-2 text-xs font-bold uppercase tracking-widest">
                {t('product.color')}
                {color && <span className="ml-2 font-normal normal-case text-muted-foreground">{color}</span>}
              </p>
              <div className="flex flex-wrap gap-2">
                {colors.map((c) => {
                  const out = !colorInStock(c)
                  const sel = color === c
                  return (
                    <button
                      key={c}
                      type="button"
                      disabled={out}
                      onClick={() => setColor(c)}
                      aria-label={c}
                      aria-pressed={sel}
                      title={c}
                      className={cn(
                        'flex items-center gap-2 rounded-full border py-1 pl-1 pr-3 text-sm transition-colors',
                        out && 'cursor-not-allowed opacity-40',
                        sel ? 'border-primary' : 'border-input hover:border-primary',
                      )}
                    >
                      <span
                        className="size-6 rounded-full border border-border"
                        style={{ backgroundColor: swatchColor(c) }}
                      />
                      {c}
                    </button>
                  )
                })}
              </div>
            </div>
          )}

          {/* เลือกไซซ์ */}
          <div className="mt-6">
            <p className="mb-2 text-xs font-bold uppercase tracking-widest">{t('product.size')}</p>
            <div className="flex flex-wrap gap-2">
              {sizes.map((s) => {
                const out = !sizeInStock(s)
                const sel = size === s
                return (
                  <button
                    key={s}
                    type="button"
                    disabled={out}
                    onClick={() => setSize(s)}
                    aria-pressed={sel}
                    className={cn(
                      'border px-4 py-2 text-sm font-medium transition-colors',
                      out && 'cursor-not-allowed text-muted-foreground/40 line-through',
                      sel ? 'border-primary bg-primary text-primary-foreground' : 'border-input hover:border-primary',
                    )}
                  >
                    {s}
                  </button>
                )
              })}
            </div>
          </div>

          <Button
            className="mt-8 w-full"
            size="lg"
            disabled={!data.in_stock || !selected || selected.stock_quantity <= 0 || addToCartMut.isPending}
            onClick={addToCart}
          >
            {addToCartMut.isPending && <Spinner />}
            {!data.in_stock
              ? t('product.outOfStock')
              : !selected
                ? t('product.selectOption')
                : t('product.addToCart')}
          </Button>
          {notice && (
            <p className="mt-3 flex flex-wrap items-center gap-2 rounded-md bg-accent/10 px-3 py-2 text-sm text-accent">
              <span>{notice}</span>
              <Link to="/cart" className="font-semibold underline">
                {t('product.viewCart')}
              </Link>
            </p>
          )}

          <div className="mt-8 border-t border-border pt-6">
            <p className="text-sm leading-relaxed text-muted-foreground">{data.description}</p>
          </div>
        </div>
      </div>

      {zoomOpen && activeImage && (
        <ImageLightbox src={activeImage.url} alt={data.name} onClose={() => setZoomOpen(false)} />
      )}
    </div>
  )
}
