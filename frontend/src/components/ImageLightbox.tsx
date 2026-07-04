import { useEffect, useState } from 'react'
import { IconClose } from '@/components/icons'
import { useLang } from '@/i18n/LanguageContext'
import { cn } from '@/lib/utils'

// ดูรูปสินค้าแบบเต็มจอ — กดที่รูปเพื่อสลับซูมเข้า/ออก (ตอนซูมเลื่อนแพนดูได้), กดพื้นหลัง/Esc เพื่อปิด
export function ImageLightbox({ src, alt, onClose }: { src: string; alt: string; onClose: () => void }) {
  const { t } = useLang()
  const [zoomed, setZoomed] = useState(false)

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKey)
    // ล็อกไม่ให้หน้าเบื้องหลังเลื่อนระหว่างเปิด lightbox
    const prevOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      document.removeEventListener('keydown', onKey)
      document.body.style.overflow = prevOverflow
    }
  }, [onClose])

  return (
    <div
      className="fixed inset-0 z-50 bg-black/90"
      role="dialog"
      aria-modal="true"
      aria-label={alt}
      onClick={onClose}
    >
      <button
        type="button"
        aria-label={t('common.close')}
        onClick={onClose}
        className="absolute right-4 top-4 z-10 text-white/80 transition-colors hover:text-white"
      >
        <IconClose className="size-7" />
      </button>

      <div className={cn('h-full w-full', zoomed ? 'overflow-auto' : 'flex items-center justify-center p-4')}>
        <img
          src={src}
          alt={alt}
          draggable={false}
          onClick={(e) => {
            e.stopPropagation() // กดที่รูป = สลับซูม ไม่ให้ทะลุไปปิด lightbox
            setZoomed((v) => !v)
          }}
          className={cn(
            'select-none',
            zoomed
              ? 'w-[160%] max-w-none cursor-zoom-out md:w-[120%]' // ใหญ่กว่าจอ → เลื่อนแพนดูได้
              : 'max-h-full max-w-full cursor-zoom-in object-contain', // พอดีจอ
          )}
        />
      </div>
    </div>
  )
}
