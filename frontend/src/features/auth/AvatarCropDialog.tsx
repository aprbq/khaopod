import { useEffect, useRef, useState } from 'react'
import type { PointerEvent as ReactPointerEvent } from 'react'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { useLang } from '@/i18n/LanguageContext'

// ขนาด viewport ที่ใช้ครอบรูป (px) — ผูกกับคณิตของ crop จึงเป็นค่าคงที่ ไม่ใช่ responsive
const VIEW = 288
// ขนาดรูปจัตุรัสที่ export ไปอัปโหลด
const OUT = 512
const MAX_ZOOM = 3

interface AvatarCropDialogProps {
  src: string // object URL ของไฟล์ที่ผู้ใช้เพิ่งเลือก
  busy: boolean // กำลังอัปโหลด — ล็อกปุ่ม/ปิด dialog ไม่ได้
  error: string
  onConfirm: (file: File) => void
  onCancel: () => void
}

// dialog ครอบรูปโปรไฟล์: ลากจัดตำแหน่ง + ซูม แล้ว export เป็น JPEG จัตุรัสผ่าน canvas
// ผู้เรียกคุมการเปิด/ปิดด้วยการ mount/unmount (แบบเดียวกับ ConfirmDialog)
export function AvatarCropDialog({ src, busy, error, onConfirm, onCancel }: AvatarCropDialogProps) {
  const { t } = useLang()
  const dialogRef = useRef<HTMLDialogElement>(null)
  const imgRef = useRef<HTMLImageElement>(null)
  const [nat, setNat] = useState<{ w: number; h: number } | null>(null) // ขนาดจริงของรูป
  const [zoom, setZoom] = useState(1)
  const [offset, setOffset] = useState({ x: 0, y: 0 })
  const drag = useRef<{ startX: number; startY: number; ox: number; oy: number } | null>(null)

  useEffect(() => {
    dialogRef.current?.showModal()
  }, [])

  // scale เริ่มต้นให้รูปพอดีด้านสั้น (cover) — zoom คูณทับจากตรงนั้น
  const baseScale = nat ? VIEW / Math.min(nat.w, nat.h) : 1
  const scale = baseScale * zoom

  // กันรูปหลุดขอบ viewport — เลื่อนได้ไม่เกินครึ่งของส่วนที่ล้นออกมา
  const clampOffset = (o: { x: number; y: number }, s: number) => {
    if (!nat) return o
    const maxX = Math.max(0, (nat.w * s - VIEW) / 2)
    const maxY = Math.max(0, (nat.h * s - VIEW) / 2)
    return { x: Math.min(maxX, Math.max(-maxX, o.x)), y: Math.min(maxY, Math.max(-maxY, o.y)) }
  }

  const onPointerDown = (e: ReactPointerEvent<HTMLDivElement>) => {
    e.currentTarget.setPointerCapture(e.pointerId)
    drag.current = { startX: e.clientX, startY: e.clientY, ox: offset.x, oy: offset.y }
  }
  const onPointerMove = (e: ReactPointerEvent<HTMLDivElement>) => {
    if (!drag.current) return
    const { startX, startY, ox, oy } = drag.current
    setOffset(clampOffset({ x: ox + e.clientX - startX, y: oy + e.clientY - startY }, scale))
  }
  const onPointerUp = () => {
    drag.current = null
  }

  const onZoom = (z: number) => {
    setZoom(z)
    setOffset((o) => clampOffset(o, baseScale * z))
  }

  // ตัดส่วนที่อยู่ใน viewport ตามตำแหน่ง/ซูมปัจจุบัน → JPEG จัตุรัส
  const confirm = async () => {
    const img = imgRef.current
    if (!img || !nat) return
    const srcSize = VIEW / scale
    const sx = nat.w / 2 - offset.x / scale - srcSize / 2
    const sy = nat.h / 2 - offset.y / scale - srcSize / 2

    const canvas = document.createElement('canvas')
    canvas.width = OUT
    canvas.height = OUT
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    // รองพื้นขาวกัน PNG/WebP ที่โปร่งใสกลายเป็นพื้นดำตอนแปลงเป็น JPEG
    ctx.fillStyle = '#fff'
    ctx.fillRect(0, 0, OUT, OUT)
    ctx.drawImage(img, sx, sy, srcSize, srcSize, 0, 0, OUT, OUT)

    const blob = await new Promise<Blob | null>((r) => canvas.toBlob(r, 'image/jpeg', 0.9))
    if (!blob) return
    onConfirm(new File([blob], 'avatar.jpg', { type: 'image/jpeg' }))
  }

  return (
    <dialog
      ref={dialogRef}
      onClose={onCancel}
      onCancel={(e) => busy && e.preventDefault()} // กัน Esc ปิดกลางทางระหว่างอัปโหลด
      onClick={(e) => {
        if (e.target === e.currentTarget && !busy) onCancel()
      }}
      className="m-auto w-[calc(100%-2rem)] max-w-sm rounded-md border border-border bg-background p-0 text-foreground backdrop:bg-black/50"
    >
      <div className="p-6">
        <h2 className="font-display text-lg uppercase">{t('profile.changeAvatar')}</h2>
        <p className="mt-1 text-sm text-muted-foreground">{t('profile.cropHint')}</p>

        <div
          className="relative mx-auto mt-4 touch-none overflow-hidden bg-muted"
          style={{ width: VIEW, height: VIEW, cursor: drag.current ? 'grabbing' : 'grab' }}
          onPointerDown={onPointerDown}
          onPointerMove={onPointerMove}
          onPointerUp={onPointerUp}
          onPointerCancel={onPointerUp}
        >
          {!nat && (
            <div className="grid h-full place-items-center">
              <Spinner className="size-6 text-muted-foreground" />
            </div>
          )}
          <img
            ref={imgRef}
            src={src}
            alt=""
            draggable={false}
            onLoad={(e) =>
              setNat({ w: e.currentTarget.naturalWidth, h: e.currentTarget.naturalHeight })
            }
            className="absolute left-1/2 top-1/2 max-w-none select-none"
            style={{
              width: nat ? nat.w * scale : undefined,
              transform: `translate(calc(-50% + ${offset.x}px), calc(-50% + ${offset.y}px))`,
              visibility: nat ? 'visible' : 'hidden',
            }}
          />
          {/* mask วงกลมให้เห็นว่ารูปโปรไฟล์จะถูกตัดตรงไหน */}
          <div
            aria-hidden
            className="pointer-events-none absolute inset-0 rounded-full shadow-[0_0_0_9999px_rgba(0,0,0,0.45)]"
          />
        </div>

        <label className="mt-4 flex items-center gap-3 text-sm text-muted-foreground">
          {t('profile.zoom')}
          <input
            type="range"
            min={1}
            max={MAX_ZOOM}
            step={0.01}
            value={zoom}
            disabled={!nat || busy}
            onChange={(e) => onZoom(Number(e.target.value))}
            className="flex-1 accent-accent"
          />
        </label>

        {error && <p className="mt-3 text-sm text-destructive">{error}</p>}

        <div className="mt-5 flex justify-end gap-3">
          <Button variant="outline" disabled={busy} onClick={onCancel}>
            {t('common.cancel')}
          </Button>
          <Button disabled={!nat || busy} onClick={() => void confirm()}>
            {busy && <Spinner />}
            {t('common.confirm')}
          </Button>
        </div>
      </div>
    </dialog>
  )
}
