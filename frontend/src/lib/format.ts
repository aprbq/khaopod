// จัดรูปแบบราคาเป็นเงินบาท
export function formatBaht(n: number): string {
  return '฿' + new Intl.NumberFormat('th-TH').format(n)
}

export function formatPriceRange(min: number, max: number): string {
  return min === max ? formatBaht(min) : `${formatBaht(min)} – ${formatBaht(max)}`
}
