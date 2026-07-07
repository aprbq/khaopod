import { useEffect, useRef } from 'react'
import { Button } from '@/components/ui/button'

interface ConfirmDialogProps {
  title: string
  desc: string
  confirmLabel: string
  cancelLabel: string
  // สีปุ่มยืนยันตามความหมายของ action — ลบ/ปฏิเสธ = destructive (ค่าเริ่มต้น), อนุมัติ = default
  confirmVariant?: 'destructive' | 'default'
  onConfirm: () => void
  onCancel: () => void
}

// กล่องยืนยันก่อนทำ action ที่ย้อนกลับไม่ได้ (เช่น ลบของออกจากตะกร้า)
// ใช้ native <dialog> เพื่อได้ focus trap + ปิดด้วย Esc โดยไม่ต้องพึ่ง library เพิ่ม
// ฝั่งผู้เรียกคุมการเปิด/ปิดด้วยการ mount/unmount คอมโพเนนต์นี้
export function ConfirmDialog({
  title,
  desc,
  confirmLabel,
  cancelLabel,
  confirmVariant = 'destructive',
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  const ref = useRef<HTMLDialogElement>(null)

  useEffect(() => {
    ref.current?.showModal()
  }, [])

  return (
    <dialog
      ref={ref}
      onClose={onCancel}
      // คลิกที่ backdrop = ยกเลิก (คลิกในเนื้อหาจะโดน div ด้านในรับไว้ก่อน)
      onClick={(e) => {
        if (e.target === e.currentTarget) onCancel()
      }}
      className="m-auto w-[calc(100%-2rem)] max-w-sm rounded-md border border-border bg-background p-0 text-foreground backdrop:bg-black/50"
    >
      <div className="p-6">
        <h2 className="font-display text-lg uppercase">{title}</h2>
        <p className="mt-2 text-sm text-muted-foreground">{desc}</p>
        <div className="mt-6 flex justify-end gap-3">
          <Button variant="outline" onClick={onCancel}>
            {cancelLabel}
          </Button>
          <Button variant={confirmVariant} onClick={onConfirm}>
            {confirmLabel}
          </Button>
        </div>
      </div>
    </dialog>
  )
}
