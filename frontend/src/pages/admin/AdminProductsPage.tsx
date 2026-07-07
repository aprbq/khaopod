import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { ConfirmDialog } from '@/components/ConfirmDialog'
import { useAdminDeleteProduct, useAdminProducts } from '@/features/admin/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { ApiError } from '@/lib/apiClient'
import { formatPriceRange } from '@/lib/format'

export function AdminProductsPage() {
  const { t } = useLang()
  useDocumentTitle(t('admin.products'))
  const { data: products, isLoading } = useAdminProducts()
  const deleteProduct = useAdminDeleteProduct()
  const [confirmId, setConfirmId] = useState<number | null>(null)
  const [err, setErr] = useState('')

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="font-display text-2xl uppercase md:text-3xl">{t('admin.products')}</h1>
        <Link to="/admin/products/new">
          <Button>+ {t('admin.addProduct')}</Button>
        </Link>
      </div>

      {err && (
        <p className="mt-4 rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{err}</p>
      )}

      {isLoading ? (
        <div className="grid min-h-[30vh] place-items-center">
          <Spinner className="size-6 text-muted-foreground" />
        </div>
      ) : !products || products.length === 0 ? (
        <p className="mt-8 text-sm text-muted-foreground">{t('admin.noProducts')}</p>
      ) : (
        <div className="mt-4 overflow-x-auto border border-border bg-background">
          <table className="w-full min-w-[46rem] text-left text-sm">
            <thead className="border-b border-border text-xs uppercase tracking-wide text-muted-foreground">
              <tr>
                <th className="px-4 py-3">{t('admin.colProduct')}</th>
                <th className="px-4 py-3">{t('admin.colCategory')}</th>
                <th className="px-4 py-3">{t('admin.colPrice')}</th>
                <th className="px-4 py-3 text-right">{t('admin.colStock')}</th>
                <th className="px-4 py-3">{t('admin.colStatus')}</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {products.map((p) => {
                const stock = p.variants.reduce((n, v) => n + v.stock_quantity, 0)
                return (
                  <tr key={p.id} className="transition-colors hover:bg-muted/50">
                    <td className="px-4 py-3">
                      <Link
                        to={`/admin/products/${p.id}`}
                        className="flex items-center gap-3 font-medium hover:underline"
                      >
                        <span className="size-10 shrink-0 overflow-hidden bg-muted">
                          {p.primary_image && (
                            <img
                              src={p.primary_image}
                              alt=""
                              loading="lazy"
                              className="h-full w-full object-cover"
                            />
                          )}
                        </span>
                        <span className="min-w-0">
                          {p.name}
                          {p.is_featured && (
                            <Badge variant="accent" className="ml-2">
                              ★
                            </Badge>
                          )}
                        </span>
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">{p.category?.name ?? '—'}</td>
                    <td className="whitespace-nowrap px-4 py-3">
                      {formatPriceRange(p.price_range.min, p.price_range.max)}
                    </td>
                    <td className="px-4 py-3 text-right">{stock}</td>
                    <td className="px-4 py-3">
                      <Badge variant={p.is_active ? 'accent' : 'outline'}>
                        {p.is_active ? t('admin.activeBadge') : t('admin.inactiveBadge')}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        type="button"
                        className="text-xs text-muted-foreground underline hover:text-destructive disabled:opacity-40"
                        disabled={deleteProduct.isPending}
                        onClick={() => setConfirmId(p.id)}
                      >
                        {t('cart.remove')}
                      </button>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}

      {confirmId !== null && (
        <ConfirmDialog
          title={t('admin.deleteProductConfirm')}
          desc={t('admin.deleteProductDesc')}
          confirmLabel={t('cart.remove')}
          cancelLabel={t('common.cancel')}
          onCancel={() => setConfirmId(null)}
          onConfirm={() => {
            const id = confirmId
            setConfirmId(null)
            setErr('')
            deleteProduct.mutate(id, {
              onError: (e) => setErr(e instanceof ApiError ? e.message : t('common.error')),
            })
          }}
        />
      )}
    </div>
  )
}
