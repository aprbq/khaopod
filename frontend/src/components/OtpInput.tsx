import { useRef } from 'react'
import { cn } from '@/lib/utils'

interface OtpInputProps {
  value: string
  onChange: (value: string) => void
  onComplete?: (value: string) => void // เรียกเมื่อกรอกครบทุกหลัก (ไว้ auto-submit)
  length?: number
  disabled?: boolean
  invalid?: boolean
  ariaLabel?: string
}

// ช่องกรอก OTP แบบแยกทีละหลัก — เลื่อนช่องอัตโนมัติ, backspace ถอยหลัง, วาง (paste) ได้
export function OtpInput({
  value,
  onChange,
  onComplete,
  length = 6,
  disabled,
  invalid,
  ariaLabel = 'OTP',
}: OtpInputProps) {
  const inputs = useRef<Array<HTMLInputElement | null>>([])

  const focusAt = (i: number) => {
    const el = inputs.current[Math.max(0, Math.min(length - 1, i))]
    el?.focus()
    el?.select()
  }

  const commit = (next: string) => {
    const clean = next.replace(/\D/g, '').slice(0, length)
    onChange(clean)
    if (clean.length === length) onComplete?.(clean)
    return clean
  }

  const handleChange = (i: number, raw: string) => {
    const digits = raw.replace(/\D/g, '')
    if (!digits) return
    // แทนที่ตำแหน่ง i ด้วยตัวเลข (พิมพ์ทีละตัว) — ถ้าเกินก็ต่อท้าย
    const next = commit(value.slice(0, i) + digits + value.slice(i + 1))
    focusAt(digits.length > 1 ? next.length : i + 1)
  }

  const handleKeyDown = (i: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace') {
      e.preventDefault()
      if (value[i]) {
        commit(value.slice(0, i) + value.slice(i + 1)) // ลบตัวที่ช่องนี้
        focusAt(i)
      } else {
        commit(value.slice(0, i - 1) + value.slice(i)) // ช่องว่าง → ลบตัวก่อนหน้า
        focusAt(i - 1)
      }
    } else if (e.key === 'ArrowLeft') {
      e.preventDefault()
      focusAt(i - 1)
    } else if (e.key === 'ArrowRight') {
      e.preventDefault()
      focusAt(i + 1)
    }
  }

  const handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault()
    const next = commit(e.clipboardData.getData('text'))
    focusAt(next.length)
  }

  return (
    <div className="flex justify-center gap-2 sm:gap-3" role="group" aria-label={ariaLabel}>
      {Array.from({ length }, (_, i) => (
        <input
          key={i}
          ref={(el) => {
            inputs.current[i] = el
          }}
          type="text"
          inputMode="numeric"
          autoComplete={i === 0 ? 'one-time-code' : 'off'}
          autoFocus={i === 0}
          maxLength={1}
          disabled={disabled}
          aria-label={`${ariaLabel} ${i + 1}`}
          aria-invalid={invalid}
          value={value[i] ?? ''}
          onChange={(e) => handleChange(i, e.target.value)}
          onKeyDown={(e) => handleKeyDown(i, e)}
          onPaste={handlePaste}
          onFocus={(e) => e.target.select()}
          className={cn(
            'size-12 rounded-md border bg-background text-center text-xl font-semibold tabular-nums outline-none transition-colors',
            'focus:border-accent focus:ring-2 focus:ring-accent/30',
            'disabled:cursor-not-allowed disabled:opacity-50',
            invalid ? 'border-destructive' : value[i] ? 'border-foreground' : 'border-input',
          )}
        />
      ))}
    </div>
  )
}
