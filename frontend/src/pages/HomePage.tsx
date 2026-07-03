import { Link } from 'react-router-dom'
import heroUrl from '@/asset/khaopod_background.jpg'
import { ProductGrid } from '@/components/ProductGrid'
import { Button } from '@/components/ui/button'
import { IconArrowRight } from '@/components/icons'
import { useFeaturedProducts } from '@/features/products/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'

function Hero() {
  const { t } = useLang()
  return (
    <section className="relative">
      <img src={heroUrl} alt="กองบัญชาการข่าวปด" className="block w-full" />
      <div className="absolute inset-x-0 bottom-4 flex justify-center md:bottom-10">
        <Link to="/shop">
          <Button variant="accent" size="lg" className="shadow-lg">
            {t('home.shopNow')}
            <IconArrowRight className="size-4" />
          </Button>
        </Link>
      </div>
    </section>
  )
}

function SectionHeader({ title, href }: { title: string; href: string }) {
  const { t } = useLang()
  return (
    <div className="mb-8 flex items-end justify-between">
      <h2 className="font-display text-2xl uppercase md:text-3xl">{title}</h2>
      <Link
        to={href}
        className="text-xs font-bold uppercase tracking-widest text-muted-foreground hover:text-foreground"
      >
        {t('common.viewAll')}
      </Link>
    </div>
  )
}

export function HomePage() {
  useDocumentTitle()
  const { t } = useLang()
  const featured = useFeaturedProducts()

  return (
    <>
      <Hero />
      <section className="mx-auto max-w-6xl px-4 py-16">
        <SectionHeader title={t('home.new')} href="/shop" />
        <ProductGrid
          products={featured.data}
          isLoading={featured.isLoading}
          isError={featured.isError}
        />
      </section>
    </>
  )
}
