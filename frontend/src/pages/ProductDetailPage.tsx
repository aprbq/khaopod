import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { useProduct } from '@/features/products/hooks'
import { formatBaht } from '@/lib/format'
import { cn } from '@/lib/utils'
import type { ProductVariant } from '@/types/product'

export function ProductDetailPage() {
  const { slug } = useParams<{ slug: string }>()
  const { data, isLoading, isError } = useProduct(slug ?? '')
  const [variant, setVariant] = useState<ProductVariant | null>(null)
  const [notice, setNotice] = useState('')

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
        <p className="text-sm text-muted-foreground">ไม่พบสินค้านี้</p>
        <Link to="/shop" className="mt-4 inline-block text-sm font-semibold underline">
          กลับไปหน้าสินค้า
        </Link>
      </div>
    )
  }

  const addToCart = () => {
    // ระบบตะกร้าจะต่อกับ backend /cart ในเฟสถัดไป
    setNotice('ระบบตะกร้ากำลังจะมาเร็ว ๆ นี้ — ตอนนี้ยังเป็นเดโมหน้าร้าน')
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      <Link to="/shop" className="text-xs font-semibold uppercase tracking-widest text-muted-foreground hover:text-foreground">
        ← สินค้าทั้งหมด
      </Link>

      <div className="mt-6 grid gap-8 md:grid-cols-2">
        {/* รูปสินค้า (placeholder) */}
        <div className="relative aspect-square bg-muted">
          <div className="flex h-full w-full items-center justify-center p-8">
            <span className="text-center text-base font-semibold uppercase tracking-widest text-muted-foreground/50">
              {data.name}
            </span>
          </div>
          {!data.in_stock && (
            <div className="absolute left-0 top-0 p-4">
              <Badge>Sold out</Badge>
            </div>
          )}
        </div>

        {/* รายละเอียด */}
        <div>
          <p className="text-xs uppercase tracking-widest text-muted-foreground">{data.category}</p>
          <h1 className="mt-2 font-display text-2xl uppercase md:text-3xl">{data.name}</h1>
          <p className="mt-4 text-xl font-bold">
            {formatBaht(variant ? variant.price : data.price_range.min)}
          </p>

          {/* เลือกตัวเลือก */}
          <div className="mt-6">
            <p className="mb-2 text-xs font-bold uppercase tracking-widest">ตัวเลือก</p>
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
            {!data.in_stock ? 'สินค้าหมด' : !variant ? 'เลือกตัวเลือกก่อน' : 'เพิ่มลงตะกร้า'}
          </Button>
          {notice && (
            <p className="mt-3 rounded-md bg-accent/10 px-3 py-2 text-sm text-accent">{notice}</p>
          )}

          <div className="mt-8 border-t border-border pt-6">
            <p className="text-sm leading-relaxed text-muted-foreground">{data.description}</p>
          </div>
        </div>
      </div>
    </div>
  )
}
