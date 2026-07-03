// แถบประกาศด้านบนแบบ ticker วิ่ง (ลุคร้านสตรีทแวร์)
const ITEMS = [
  'จัดส่งฟรีเมื่อซื้อครบ ฿1,000',
  'สินค้าลิขสิทธิ์แท้จากเพจกองบัญชาการข่าวปด',
  'ของใหม่ทุกสัปดาห์',
]

export function AnnouncementBar() {
  const line = ITEMS.join('  •  ')
  const block = `${line}  •  `
  return (
    <div className="overflow-hidden bg-primary text-primary-foreground">
      <div className="flex w-max animate-marquee py-2 text-[11px] font-semibold uppercase tracking-widest whitespace-nowrap">
        <span className="px-2">{block}</span>
        <span className="px-2" aria-hidden="true">
          {block}
        </span>
      </div>
    </div>
  )
}
