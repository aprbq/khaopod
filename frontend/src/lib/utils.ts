import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

// รวม class แบบมีเงื่อนไข + แก้ conflict ของ tailwind (อย่าต่อ string เอง)
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs))
}
