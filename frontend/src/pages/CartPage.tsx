import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { IconBag } from '@/components/icons'

// placeholder — ระบบตะกร้าจริงจะต่อกับ backend /cart ในเฟสถัดไป
export function CartPage() {
  return (
    <div className="mx-auto grid min-h-[50vh] max-w-6xl place-items-center px-4 py-16 text-center">
      <div>
        <IconBag className="mx-auto size-10 text-muted-foreground" />
        <h1 className="mt-4 font-display text-2xl uppercase">ตะกร้าของคุณยังว่าง</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          ระบบตะกร้ากำลังจะเปิดใช้งานเร็ว ๆ นี้
        </p>
        <Link to="/shop" className="mt-6 inline-block">
          <Button size="lg">เลือกซื้อสินค้า</Button>
        </Link>
      </div>
    </div>
  )
}
