import * as React from 'react'
import { cn } from '@/lib/utils'

type Variant = 'default' | 'accent' | 'outline'

const variants: Record<Variant, string> = {
  default: 'bg-primary text-primary-foreground',
  accent: 'bg-accent text-accent-foreground',
  outline: 'border border-border text-foreground',
}

export function Badge({
  className,
  variant = 'default',
  ...props
}: React.HTMLAttributes<HTMLSpanElement> & { variant?: Variant }) {
  return (
    <span
      className={cn(
        'inline-flex items-center px-2 py-0.5 text-[10px] font-bold uppercase tracking-widest',
        variants[variant],
        className,
      )}
      {...props}
    />
  )
}
