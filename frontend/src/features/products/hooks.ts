import { useQuery } from '@tanstack/react-query'
import type { ProductQuery } from '@/types/product'
import { productApi } from './api'

// แคตตาล็อกสินค้า cache ได้นาน (ข้อมูลไม่เปลี่ยนบ่อย) — ตั้ง staleTime ที่ query client กลาง

export function useProducts(query?: ProductQuery) {
  return useQuery({
    // query key รวม params เพื่อ cache แยกตาม filter/หมวด
    queryKey: ['products', query ?? {}],
    queryFn: () => productApi.list(query),
  })
}

// สินค้าแนะนำ — backend ยังไม่มี filter featured จึงดึงทั้งหมดแล้วกรองฝั่ง client
export function useFeaturedProducts() {
  return useQuery({
    queryKey: ['products', 'featured'],
    queryFn: async () => {
      const products = await productApi.list()
      return products.filter((p) => p.is_featured)
    },
  })
}

export function useProduct(slug: string) {
  return useQuery({
    queryKey: ['product', slug],
    queryFn: () => productApi.get(slug),
    enabled: slug !== '', // อย่ายิงถ้ายังไม่มี slug
  })
}
