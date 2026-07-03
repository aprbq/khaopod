import { useSearchParams } from 'react-router-dom'
import { ProductGrid } from '@/components/ProductGrid'
import { useProducts } from '@/features/products/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'

export function ShopPage() {
  const { t } = useLang()
  const [params] = useSearchParams()
  const category = params.get('category')
  const { data, isLoading, isError } = useProducts()

  // กรองตามหมวด ถ้ามี ?category= ใน URL (ชื่อหมวดยังมาจากข้อมูลสินค้าโดยตรง)
  const products = category ? data?.filter((p) => p.category === category) : data
  const heading = category ?? t('shop.collection')
  useDocumentTitle(heading)

  return (
    <section className="mx-auto max-w-6xl px-4 py-12">
      <div className="mb-8 border-b border-border pb-6">
        <h1 className="font-display text-3xl uppercase md:text-4xl">{heading}</h1>
        {products && (
          <p className="mt-1 text-sm text-muted-foreground">
            {t('common.items', { n: products.length })}
          </p>
        )}
      </div>
      <ProductGrid products={products} isLoading={isLoading} isError={isError} />
    </section>
  )
}
