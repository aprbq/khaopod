import { Badge } from '@/components/ui/badge'
import { useLang } from '@/i18n/LanguageContext'
import type { TranslationKey } from '@/i18n/translations'
import type { OrderStatus } from '@/types/order'

// badge สถานะออเดอร์ — โทนสีตามความหมาย: pending เด่น, จบแล้ว (ยกเลิก/คืนเงิน) จาง
export function OrderStatusBadge({ status }: { status: OrderStatus }) {
  const { t } = useLang()
  const variant =
    status === 'cancelled' || status === 'refunded'
      ? 'outline'
      : status === 'pending'
        ? 'default'
        : 'accent'
  return <Badge variant={variant}>{t(`orderStatus.${status}` as TranslationKey)}</Badge>
}
