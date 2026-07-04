import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ImageLightbox } from '@/components/ImageLightbox'
import { Spinner } from '@/components/ui/spinner'
import { useProduct } from '@/features/products/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { formatBaht } from '@/lib/format'
import { cn } from '@/lib/utils'
import type { ProductVariant } from '@/types/product'

export function ProductDetailPage() {
  const { t } = useLang()
  const { slug } = useParams<{ slug: string }>()
  const { data, isLoading, isError } = useProduct(slug ?? '')
  const [variant, setVariant] = useState<ProductVariant | null>(null)
  const [activeIdx, setActiveIdx] = useState(0)
  const [zoomOpen, setZoomOpen] = useState(false)
  const [notice, setNotice] = useState('')
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
    // ระบบตะกร้าจะต่อกับ backend /cart ในเฟสถัดไป
    setNotice(t('product.cartComingSoon'))
  }

  const images = data.images ?? []
  const activeImage = images[activeIdx]

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
            {formatBaht(variant ? variant.price : data.price_range.min)}
          </p>

          {/* เลือกตัวเลือก */}
          <div className="mt-6">
            <p className="mb-2 text-xs font-bold uppercase tracking-widest">{t('product.options')}</p>
            <div className="flex flex-wrap gap-2">
              {data.variants.map((v) => {
                const out = v.stock_quantity <= 0
                const selected = variant?.id === v.id
                return (
                  <button
                    key={v.id}
                    type="button"
                    disabled={out}
                    onClick={() => setVariant(v)}
                    className={cn(
                      'border px-4 py-2 text-sm font-medium transition-colors',
                      out && 'cursor-not-allowed text-muted-foreground/40 line-through',
                      selected ? 'border-primary bg-primary text-primary-foreground' : 'border-input hover:border-primary',
                    )}
                  >
                    {v.variant_name}
                  </button>
                )
              })}
            </div>
          </div>

          <Button
            className="mt-8 w-full"
            size="lg"
            disabled={!data.in_stock || !variant}
            onClick={addToCart}
          >
            {!data.in_stock
              ? t('product.outOfStock')
              : !variant
                ? t('product.selectOption')
                : t('product.addToCart')}
          </Button>
          {notice && (
            <p className="mt-3 rounded-md bg-accent/10 px-3 py-2 text-sm text-accent">{notice}</p>
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
