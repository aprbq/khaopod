// จัดรูปแบบราคาเป็นเงินบาท
export function formatBaht(n: number): string {
  return '฿' + new Intl.NumberFormat('th-TH').format(n)
}

export function formatPriceRange(min: number, max: number): string {
  return min === max ? formatBaht(min) : `${formatBaht(min)} – ${formatBaht(max)}`
}

// วันที่แบบอ่านง่ายตามภาษา UI (th-TH ได้ปีพุทธศักราชตาม locale)
export function formatDate(iso: string, lang: 'th' | 'en' = 'th'): string {
  return new Intl.DateTimeFormat(lang === 'th' ? 'th-TH' : 'en-GB', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  }).format(new Date(iso))
}

// วันที่ + เวลา (ใช้ในหลังบ้าน เช่น login ล่าสุด)
export function formatDateTime(iso: string, lang: 'th' | 'en' = 'th'): string {
  return new Intl.DateTimeFormat(lang === 'th' ? 'th-TH' : 'en-GB', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(iso))
}
