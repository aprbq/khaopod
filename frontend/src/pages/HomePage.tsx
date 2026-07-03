import { Link } from 'react-router-dom'
import { ProductGrid } from '@/components/ProductGrid'
import { Button } from '@/components/ui/button'
import { IconArrowRight } from '@/components/icons'
import { useFeaturedProducts } from '@/features/products/hooks'

function Hero() {
  return (
    <section className="bg-primary text-primary-foreground">
      <div className="mx-auto max-w-6xl px-4 py-24 md:py-36">
        <p className="text-xs font-bold uppercase tracking-[0.3em] text-primary-foreground/60">
          Official Merch
        </p>
        <h1 className="mt-4 font-display text-6xl leading-none uppercase md:text-8xl">
          KHAOPOD<span className="text-accent">.</span>
        </h1>
        <p className="mt-6 max-w-md text-sm leading-relaxed text-primary-foreground/70">
          เสื้อผ้าและของสะสมสตรีทแวร์อย่างเป็นทางการ จากเพจกองบัญชาการข่าวปด
          ดีไซน์มินิมอล ใส่ได้จริง ของแท้ทุกชิ้น
        </p>
        <Link to="/shop" className="mt-8 inline-block">
          <Button variant="accent" size="lg">
            ช้อปเลย
            <IconArrowRight className="size-4" />
          </Button>
        </Link>
      </div>
    </section>
  )
}

function SectionHeader({ title, href }: { title: string; href: string }) {
  return (
    <div className="mb-8 flex items-end justify-between">
      <h2 className="font-display text-2xl uppercase md:text-3xl">{title}</h2>
      <Link
        to={href}
        className="text-xs font-bold uppercase tracking-widest text-muted-foreground hover:text-foreground"
      >
        ดูทั้งหมด
      </Link>
    </div>
  )
}

export function HomePage() {
  const featured = useFeaturedProducts()

  return (
    <>
      <Hero />
      <section className="mx-auto max-w-6xl px-4 py-16">
        <SectionHeader title="มาใหม่" href="/shop" />
        <ProductGrid
          products={featured.data}
          isLoading={featured.isLoading}
          isError={featured.isError}
        />
      </section>
    </>
  )
}
