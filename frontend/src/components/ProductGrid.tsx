import { ProductCard } from '@/components/ProductCard'
import { Spinner } from '@/components/ui/spinner'
import type { Product } from '@/types/product'

interface Props {
  products: Product[] | undefined
  isLoading?: boolean
  isError?: boolean
}

export function ProductGrid({ products, isLoading, isError }: Props) {
  if (isLoading) {
    return (
      <div className="grid place-items-center py-24">
        <Spinner className="size-6 text-muted-foreground" />
      </div>
    )
  }
  if (isError) {
    return (
      <p className="py-24 text-center text-sm text-destructive">
        โหลดสินค้าไม่สำเร็จ กรุณาลองใหม่อีกครั้ง
      </p>
    )
  }
  if (!products || products.length === 0) {
    return <p className="py-24 text-center text-sm text-muted-foreground">ยังไม่มีสินค้า</p>
  }

  return (
    <div className="grid grid-cols-2 gap-x-4 gap-y-8 md:grid-cols-3 lg:grid-cols-4">
      {products.map((p) => (
        <ProductCard key={p.id} product={p} />
      ))}
    </div>
  )
}
