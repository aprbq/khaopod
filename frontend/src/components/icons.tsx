// ไอคอน inline SVG (stroke-based) — ไม่พึ่ง lib ภายนอก, ใช้ currentColor
type IconProps = { className?: string }

const base = {
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 1.7,
  strokeLinecap: 'round' as const,
  strokeLinejoin: 'round' as const,
}

export function IconMenu({ className }: IconProps) {
  return (
    <svg {...base} className={className} width="24" height="24" aria-hidden="true">
      <path d="M3 6h18M3 12h18M3 18h18" />
    </svg>
  )
}

export function IconClose({ className }: IconProps) {
  return (
    <svg {...base} className={className} width="24" height="24" aria-hidden="true">
      <path d="M6 6l12 12M18 6L6 18" />
    </svg>
  )
}

export function IconBag({ className }: IconProps) {
  return (
    <svg {...base} className={className} width="24" height="24" aria-hidden="true">
      <path d="M6 8h12l-1 12H7L6 8Z" />
      <path d="M9 8V6a3 3 0 0 1 6 0v2" />
    </svg>
  )
}

export function IconUser({ className }: IconProps) {
  return (
    <svg {...base} className={className} width="24" height="24" aria-hidden="true">
      <circle cx="12" cy="8" r="3.5" />
      <path d="M5 20a7 7 0 0 1 14 0" />
    </svg>
  )
}

export function IconArrowRight({ className }: IconProps) {
  return (
    <svg {...base} className={className} width="24" height="24" aria-hidden="true">
      <path d="M5 12h14M13 6l6 6-6 6" />
    </svg>
  )
}
