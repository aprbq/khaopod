import { Link } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { formatPriceRange } from '@/lib/format'
import { cn } from '@/lib/utils'
import type { Product } from '@/types/product'

export function ProductCard({ product }: { product: Product }) {
  const { slug, name, primary_image, in_stock, is_featured, category, price_range } = product

  return (
    <Link to={`/product/${slug}`} className="group block">
      <div className="relative aspect-square overflow-hidden bg-muted">
        {primary_image ? (
          <img
            src={primary_image}
            alt={name}
            loading="lazy"
            className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
          />
        ) : (
          // placeholder แบบ typographic (ยังไม่มีรูปสินค้าจริง)
          <div className="flex h-full w-full items-center justify-center p-6">
            <span className="text-center text-sm font-semibold uppercase leading-snug tracking-widest text-muted-foreground/50">
              {name}
            </span>
          </div>
        )}

        <div className="absolute left-0 top-0 flex flex-col gap-1 p-3">
          {is_featured && in_stock && <Badge variant="accent">New</Badge>}
          {!in_stock && <Badge variant="default">Sold out</Badge>}
        </div>
      </div>

      <div className="flex items-start justify-between gap-2 py-3">
        <div className="min-w-0">
          <p className="truncate text-sm font-semibold group-hover:underline">{name}</p>
          <p className="text-[11px] uppercase tracking-wider text-muted-foreground">
            {category?.name}
          </p>
        </div>
        <p
          className={cn(
            'shrink-0 text-sm font-bold',
            !in_stock && 'text-muted-foreground line-through',
          )}
        >
          {formatPriceRange(price_range.min, price_range.max)}
        </p>
      </div>
    </Link>
  )
}
