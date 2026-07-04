import { useSearchParams } from 'react-router-dom'
import { ProductGrid } from '@/components/ProductGrid'
import { useProducts } from '@/features/products/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'

export function ShopPage() {
  const { t } = useLang()
  const [params] = useSearchParams()
  const category = params.get('category') ?? undefined
  // กรองตามหมวดที่ backend (ส่ง slug ผ่าน ?category=) ไม่กรองฝั่ง client
  const { data, isLoading, isError } = useProducts(category ? { category } : undefined)

  // หัวเรื่อง: ใช้ชื่อหมวดจากสินค้าที่ได้กลับมา ถ้าไม่มีก็ fallback เป็น slug/ชื่อคอลเลกชัน
  const heading = category ? (data?.[0]?.category?.name ?? category) : t('shop.collection')
  useDocumentTitle(heading)

  return (
    <section className="mx-auto max-w-6xl px-4 py-12">
      <div className="mb-8 border-b border-border pb-6">
        <h1 className="font-display text-3xl uppercase md:text-4xl">{heading}</h1>
        {data && (
          <p className="mt-1 text-sm text-muted-foreground">
            {t('common.items', { n: data.length })}
          </p>
        )}
      </div>
      <ProductGrid products={data} isLoading={isLoading} isError={isError} />
    </section>
  )
}
