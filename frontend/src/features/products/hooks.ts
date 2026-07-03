import { useQuery } from '@tanstack/react-query'
import type { Product, ProductDetail } from '@/types/product'
import { mockDetail, mockProducts } from './mock'

// NOTE: ตอนนี้คืน mock — เมื่อ backend มี endpoint สินค้าแล้ว เปลี่ยน queryFn เป็น
//   apiFetch<Paginated<Product>>('/products?...')  และ apiFetch<ProductDetail>(`/products/${slug}`)
// โครง type ตรงกับ docs/rest_api.md แล้ว จึงสลับได้โดยไม่ต้องแก้ component

function delay<T>(value: T, ms = 300): Promise<T> {
  return new Promise((resolve) => setTimeout(() => resolve(value), ms))
}

export function useProducts() {
  return useQuery({
    queryKey: ['products'],
    queryFn: (): Promise<Product[]> => delay(mockProducts),
  })
}

export function useFeaturedProducts() {
  return useQuery({
    queryKey: ['products', 'featured'],
    queryFn: (): Promise<Product[]> => delay(mockProducts.filter((p) => p.is_featured)),
  })
}

export function useProduct(slug: string) {
  return useQuery({
    queryKey: ['product', slug],
    queryFn: (): Promise<ProductDetail | null> => delay(mockDetail(slug) ?? null),
  })
}
