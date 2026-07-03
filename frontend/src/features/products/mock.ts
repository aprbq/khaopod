import type { Product, ProductDetail } from '@/types/product'

// ---- ข้อมูลสินค้าจำลอง (placeholder) ----
// ใช้ชั่วคราวจนกว่า backend จะมี /products — โครงตรงกับ type จริงแล้ว สลับง่าย

export const mockProducts: Product[] = [
  {
    id: 1,
    name: 'เสื้อยืดกองบัญชาการข่าวปด',
    slug: 'kbc-tee-hq',
    base_price: 390,
    is_featured: true,
    primary_image: null,
    price_range: { min: 390, max: 420 },
    in_stock: true,
    category: 'เสื้อยืด',
  },
  {
    id: 2,
    name: 'เสื้อโอเวอร์ไซซ์ BREAKING NEWS',
    slug: 'breaking-news-oversized',
    base_price: 590,
    is_featured: true,
    primary_image: null,
    price_range: { min: 590, max: 590 },
    in_stock: true,
    category: 'เสื้อยืด',
  },
  {
    id: 3,
    name: 'ฮู้ดดี้ ข่าวปด CLASSIC',
    slug: 'khaopod-hoodie-classic',
    base_price: 890,
    is_featured: false,
    primary_image: null,
    price_range: { min: 890, max: 950 },
    in_stock: true,
    category: 'ฮู้ดดี้',
  },
  {
    id: 4,
    name: 'สติกเกอร์เซ็ต ข่าวเด่นประจำวัน',
    slug: 'sticker-daily-set',
    base_price: 120,
    is_featured: false,
    primary_image: null,
    price_range: { min: 120, max: 120 },
    in_stock: true,
    category: 'สติกเกอร์',
  },
  {
    id: 5,
    name: 'หมวกบักเก็ต PRESS',
    slug: 'press-bucket-hat',
    base_price: 450,
    is_featured: true,
    primary_image: null,
    price_range: { min: 450, max: 450 },
    in_stock: false,
    category: 'หมวก',
  },
  {
    id: 6,
    name: 'กระเป๋าผ้า TOTE ข่าวปด',
    slug: 'khaopod-tote',
    base_price: 290,
    is_featured: false,
    primary_image: null,
    price_range: { min: 290, max: 290 },
    in_stock: true,
    category: 'กระเป๋า',
  },
  {
    id: 7,
    name: 'เข็มกลัด ENAMEL PIN นักข่าว',
    slug: 'reporter-enamel-pin',
    base_price: 150,
    is_featured: false,
    primary_image: null,
    price_range: { min: 150, max: 150 },
    in_stock: true,
    category: 'ของสะสม',
  },
  {
    id: 8,
    name: 'เสื้อยืด LIMITED — ฉบับพิเศษ',
    slug: 'limited-special-tee',
    base_price: 690,
    is_featured: true,
    primary_image: null,
    price_range: { min: 690, max: 690 },
    in_stock: false,
    category: 'เสื้อยืด',
  },
]

export function mockDetail(slug: string): ProductDetail | undefined {
  const p = mockProducts.find((x) => x.slug === slug)
  if (!p) return undefined
  return {
    ...p,
    description:
      'สินค้าอย่างเป็นทางการจากเพจกองบัญชาการข่าวปด งานพิมพ์คุณภาพ ผ้าฝ้ายเนื้อดี ใส่สบาย ' +
      'ดีไซน์สตรีทแวร์มินิมอล สื่อถึงตัวตนของชาวข่าวปดแบบเท่ ๆ',
    images: [],
    variants: [
      { id: p.id * 10 + 1, variant_name: 'ไซซ์ M', price: p.price_range.min, stock_quantity: p.in_stock ? 12 : 0 },
      { id: p.id * 10 + 2, variant_name: 'ไซซ์ L', price: p.price_range.max, stock_quantity: p.in_stock ? 5 : 0 },
      { id: p.id * 10 + 3, variant_name: 'ไซซ์ XL', price: p.price_range.max, stock_quantity: 0 },
    ],
  }
}
