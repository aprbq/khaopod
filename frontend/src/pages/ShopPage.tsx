import { ProductGrid } from '@/components/ProductGrid'
import { useProducts } from '@/features/products/hooks'

export function ShopPage() {
  const { data, isLoading, isError } = useProducts()

  return (
    <section className="mx-auto max-w-6xl px-4 py-12">
      <div className="mb-8 border-b border-border pb-6">
        <h1 className="font-display text-3xl uppercase md:text-4xl">สินค้าทั้งหมด</h1>
        {data && <p className="mt-1 text-sm text-muted-foreground">{data.length} รายการ</p>}
      </div>
      <ProductGrid products={data} isLoading={isLoading} isError={isError} />
    </section>
  )
}
