import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import type { ChangeEvent } from 'react'
import { useForm } from 'react-hook-form'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { z } from 'zod'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { ConfirmDialog } from '@/components/ConfirmDialog'
import {
  useAdminAddImage,
  useAdminCreateProduct,
  useAdminCreateVariant,
  useAdminDeleteImage,
  useAdminDeleteVariant,
  useAdminProduct,
  useAdminSetPrimaryImage,
  useAdminUpdateProduct,
  useAdminUpdateVariant,
  useCategories,
} from '@/features/admin/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { ApiError } from '@/lib/apiClient'
import type { VariantInput } from '@/types/admin'
import type { ProductVariant } from '@/types/product'

// หน้าเดียวใช้ทั้งสร้าง (id = 0) และแก้ไข — ต้องบันทึกสินค้าก่อนจึงเพิ่มตัวเลือก/รูปได้
export function AdminProductFormPage() {
  const { t } = useLang()
  const params = useParams()
  const id = Number(params.id ?? 0) || 0
  const isNew = id === 0
  useDocumentTitle(isNew ? t('admin.addProduct') : t('admin.editProduct'))

  const navigate = useNavigate()
  const { data: product, isLoading } = useAdminProduct(id)
  const { data: categories } = useCategories()
  const createProduct = useAdminCreateProduct()
  const updateProduct = useAdminUpdateProduct()
  const [err, setErr] = useState('')
  const [saved, setSaved] = useState(false)

  const schema = z.object({
    name: z.string().min(1, t('addr.required')),
    slug: z.string().min(1, t('addr.required')),
    description: z.string(),
    // เก็บเป็น string ในฟอร์ม (input number ก็คืน string) แล้วแปลงตอน submit
    // — z.coerce ใช้ไม่ได้เพราะ input type เป็น unknown ชนกับ typing ของ react-hook-form
    base_price: z
      .string()
      .refine((v) => v !== '' && !Number.isNaN(Number(v)) && Number(v) >= 0, t('admin.priceInvalid')),
    category_id: z.string(), // ค่า select เป็น string ("" = ไม่จัดหมวด)
    is_active: z.boolean(),
    is_featured: z.boolean(),
  })
  type Form = z.infer<typeof schema>

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<Form>({
    resolver: zodResolver(schema),
    // ตอนแก้ไข: รอข้อมูลโหลดแล้วค่อย render ฟอร์ม (ดูเงื่อนไข isLoading ด้านล่าง) จึงใช้ values ได้ตรง ๆ
    values: product
      ? {
          name: product.name,
          slug: product.slug,
          description: product.description ?? '',
          base_price: String(product.base_price),
          category_id: product.category ? String(product.category.id) : '',
          is_active: product.is_active,
          is_featured: product.is_featured,
        }
      : undefined,
    defaultValues: {
      name: '',
      slug: '',
      description: '',
      base_price: '0',
      category_id: '',
      is_active: true,
      is_featured: false,
    },
  })

  if (!isNew && isLoading) {
    return (
      <div className="grid min-h-[40vh] place-items-center">
        <Spinner className="size-6 text-muted-foreground" />
      </div>
    )
  }
  if (!isNew && !product) {
    return (
      <div>
        <p className="text-sm text-muted-foreground">{t('grid.empty')}</p>
        <Link to="/admin/products" className="mt-2 inline-block text-sm underline">
          ← {t('admin.products')}
        </Link>
      </div>
    )
  }

  const onApiError = (e: unknown) => setErr(e instanceof ApiError ? e.message : t('common.error'))

  const onSubmit = handleSubmit((v) => {
    setErr('')
    setSaved(false)
    const input = {
      name: v.name,
      slug: v.slug,
      description: v.description || undefined,
      base_price: Number(v.base_price),
      category_id: v.category_id ? Number(v.category_id) : null,
      is_active: v.is_active,
      is_featured: v.is_featured,
    }
    if (isNew) {
      createProduct.mutate(input, {
        // สร้างเสร็จพาไปหน้าแก้ไข เพื่อเพิ่มตัวเลือก/รูปต่อได้เลย
        onSuccess: (p) => navigate(`/admin/products/${p.id}`, { replace: true }),
        onError: onApiError,
      })
    } else {
      updateProduct.mutate({ id, input }, { onSuccess: () => setSaved(true), onError: onApiError })
    }
  })

  const pending = createProduct.isPending || updateProduct.isPending

  return (
    <div>
      <Link
        to="/admin/products"
        className="text-sm text-muted-foreground underline hover:text-foreground"
      >
        ← {t('admin.products')}
      </Link>
      <h1 className="mt-3 font-display text-2xl uppercase md:text-3xl">
        {isNew ? t('admin.addProduct') : t('admin.editProduct')}
      </h1>

      {err && (
        <p className="mt-4 rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{err}</p>
      )}
      {saved && (
        <p className="mt-4 rounded-md bg-accent/10 px-3 py-2 text-sm text-accent">
          {t('profile.saved')}
        </p>
      )}

      <div className="mt-6 grid gap-6 lg:grid-cols-2">
        {/* ข้อมูลสินค้า */}
        <Card className="bg-background">
          <CardContent className="pt-6">
            <form onSubmit={onSubmit} className="flex flex-col gap-4 text-sm" noValidate>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="p-name">{t('admin.prodName')}</Label>
                <Input id="p-name" {...register('name')} />
                {errors.name && <p className="text-destructive">{errors.name.message}</p>}
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="p-slug">{t('admin.prodSlug')}</Label>
                <Input id="p-slug" placeholder="khaopod-tee" {...register('slug')} />
                {errors.slug && <p className="text-destructive">{errors.slug.message}</p>}
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="p-desc">{t('admin.prodDesc')}</Label>
                <textarea
                  id="p-desc"
                  rows={4}
                  {...register('description')}
                  className="rounded-md border border-input bg-background px-3 py-2 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="p-price">{t('admin.prodBasePrice')}</Label>
                  <Input id="p-price" type="number" step="0.01" min="0" {...register('base_price')} />
                  {errors.base_price && (
                    <p className="text-destructive">{errors.base_price.message}</p>
                  )}
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="p-cat">{t('admin.prodCategory')}</Label>
                  <select
                    id="p-cat"
                    {...register('category_id')}
                    className="h-10 rounded-md border border-input bg-background px-3 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  >
                    <option value="">{t('admin.prodNoCategory')}</option>
                    {(categories ?? []).map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.name}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
              <label className="flex items-center gap-2">
                <input type="checkbox" className="accent-accent" {...register('is_active')} />
                {t('admin.prodActive')}
              </label>
              <label className="flex items-center gap-2">
                <input type="checkbox" className="accent-accent" {...register('is_featured')} />
                {t('admin.prodFeatured')}
              </label>

              <Button type="submit" className="self-start" disabled={pending}>
                {pending && <Spinner />}
                {t('profile.save')}
              </Button>
            </form>
          </CardContent>
        </Card>

        <div className="flex flex-col gap-6">
          {isNew ? (
            <p className="rounded-md bg-muted px-4 py-3 text-sm text-muted-foreground">
              {t('admin.saveProductFirst')}
            </p>
          ) : (
            <>
              <VariantsSection
                productId={id}
                variants={product?.variants ?? []}
                onError={onApiError}
              />
              <ImagesSection productId={id} onError={onApiError} />
            </>
          )}
        </div>
      </div>
    </div>
  )
}

// ---- ตัวเลือกสินค้า (ไซซ์/สี/ราคา/สต็อก) ----

function VariantsSection({
  productId,
  variants,
  onError,
}: {
  productId: number
  variants: ProductVariant[]
  onError: (e: unknown) => void
}) {
  const { t } = useLang()
  const createVariant = useAdminCreateVariant()
  const deleteVariant = useAdminDeleteVariant()
  const [confirmId, setConfirmId] = useState<number | null>(null)
  const [adding, setAdding] = useState(false)

  return (
    <Card className="bg-background">
      <CardHeader>
        <CardTitle className="text-lg">{t('admin.variants')}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-3 text-sm">
        {variants.map((v) => (
          <VariantRow key={v.id} variant={v} onError={onError} onDelete={() => setConfirmId(v.id)} />
        ))}

        {adding ? (
          <VariantRow
            onError={onError}
            onCreate={(input) =>
              createVariant.mutate(
                { productId, input },
                { onSuccess: () => setAdding(false), onError },
              )
            }
            onCancel={() => setAdding(false)}
            pending={createVariant.isPending}
          />
        ) : (
          <button
            type="button"
            className="self-start text-sm underline hover:text-accent"
            onClick={() => setAdding(true)}
          >
            + {t('admin.addVariant')}
          </button>
        )}
      </CardContent>

      {confirmId !== null && (
        <ConfirmDialog
          title={t('admin.deleteVariantConfirm')}
          desc={t('admin.deleteVariantDesc')}
          confirmLabel={t('cart.remove')}
          cancelLabel={t('common.cancel')}
          onCancel={() => setConfirmId(null)}
          onConfirm={() => {
            const vid = confirmId
            setConfirmId(null)
            deleteVariant.mutate(vid, { onError })
          }}
        />
      )}
    </Card>
  )
}

// แถวแก้ไข/สร้างตัวเลือก — controlled input ทั้งแถว กดบันทึกค่อยยิง API
function VariantRow({
  variant,
  pending,
  onError,
  onCreate,
  onCancel,
  onDelete,
}: {
  variant?: ProductVariant
  pending?: boolean
  onError: (e: unknown) => void
  onCreate?: (input: VariantInput) => void
  onCancel?: () => void
  onDelete?: () => void
}) {
  const { t } = useLang()
  const updateVariant = useAdminUpdateVariant()
  const [name, setName] = useState(variant?.variant_name ?? '')
  const [color, setColor] = useState(variant?.color ?? '')
  const [price, setPrice] = useState(String(variant?.price ?? ''))
  const [stock, setStock] = useState(String(variant?.stock_quantity ?? 0))
  const [active, setActive] = useState(variant?.is_active ?? true)

  const toInput = (): VariantInput | null => {
    const p = Number(price)
    const s = Number(stock)
    if (!name.trim() || Number.isNaN(p) || p < 0 || !Number.isInteger(s) || s < 0) return null
    return {
      variant_name: name.trim(),
      color: color.trim() || undefined,
      sku: variant?.sku,
      price: p,
      stock_quantity: s,
      is_active: active,
    }
  }

  const save = () => {
    const input = toInput()
    if (!input) {
      onError(new ApiError('VALIDATION_ERROR', t('admin.variantInvalid'), 422))
      return
    }
    if (onCreate) {
      onCreate(input)
    } else if (variant) {
      updateVariant.mutate({ variantId: variant.id, input }, { onError })
    }
  }

  const busy = pending || updateVariant.isPending

  return (
    <div className="grid grid-cols-2 items-end gap-2 border border-border p-3 sm:grid-cols-[1fr_5rem_6rem_4.5rem_auto]">
      <label className="flex flex-col gap-1 text-xs text-muted-foreground">
        {t('admin.varName')}
        <Input value={name} onChange={(e) => setName(e.target.value)} className="h-9" />
      </label>
      <label className="flex flex-col gap-1 text-xs text-muted-foreground">
        {t('admin.varColor')}
        <Input value={color} onChange={(e) => setColor(e.target.value)} className="h-9" />
      </label>
      <label className="flex flex-col gap-1 text-xs text-muted-foreground">
        {t('admin.varPrice')}
        <Input
          type="number"
          min="0"
          step="0.01"
          value={price}
          onChange={(e) => setPrice(e.target.value)}
          className="h-9"
        />
      </label>
      <label className="flex flex-col gap-1 text-xs text-muted-foreground">
        {t('admin.varStock')}
        <Input
          type="number"
          min="0"
          step="1"
          value={stock}
          onChange={(e) => setStock(e.target.value)}
          className="h-9"
        />
      </label>
      <div className="col-span-2 flex items-center gap-3 sm:col-span-1">
        <label className="flex items-center gap-1.5 text-xs">
          <input
            type="checkbox"
            className="accent-accent"
            checked={active}
            onChange={(e) => setActive(e.target.checked)}
          />
          {t('admin.colActive')}
        </label>
        <Button size="sm" disabled={busy} onClick={save}>
          {busy && <Spinner className="size-3" />}
          {t('profile.save')}
        </Button>
        {onCancel && (
          <Button size="sm" variant="outline" disabled={busy} onClick={onCancel}>
            {t('common.cancel')}
          </Button>
        )}
        {onDelete && (
          <button
            type="button"
            className="text-xs text-muted-foreground underline hover:text-destructive"
            onClick={onDelete}
          >
            {t('cart.remove')}
          </button>
        )}
      </div>
    </div>
  )
}

// ---- รูปสินค้า ----

function ImagesSection({
  productId,
  onError,
}: {
  productId: number
  onError: (e: unknown) => void
}) {
  const { t } = useLang()
  const { data: product } = useAdminProduct(productId)
  const addImage = useAdminAddImage()
  const deleteImage = useAdminDeleteImage()
  const setPrimary = useAdminSetPrimaryImage()
  const [confirmId, setConfirmId] = useState<number | null>(null)

  const onPick = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = ''
    if (!file) return
    addImage.mutate({ productId, file }, { onError })
  }

  const images = product?.images ?? []

  return (
    <Card className="bg-background">
      <CardHeader>
        <CardTitle className="text-lg">{t('admin.images')}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-3 text-sm">
        <div className="grid grid-cols-3 gap-3 sm:grid-cols-4">
          {images.map((img) => (
            <figure key={img.id} className="relative border border-border">
              <img src={img.url} alt="" loading="lazy" className="aspect-square w-full object-cover" />
              {img.is_primary && (
                <Badge variant="accent" className="absolute left-1 top-1">
                  {t('admin.primaryBadge')}
                </Badge>
              )}
              <figcaption className="flex justify-between gap-1 p-1 text-[11px]">
                {!img.is_primary ? (
                  <button
                    type="button"
                    className="underline hover:text-accent disabled:opacity-40"
                    disabled={setPrimary.isPending}
                    onClick={() => setPrimary.mutate({ productId, imageId: img.id }, { onError })}
                  >
                    {t('admin.setPrimary')}
                  </button>
                ) : (
                  <span />
                )}
                <button
                  type="button"
                  className="text-muted-foreground underline hover:text-destructive"
                  onClick={() => setConfirmId(img.id)}
                >
                  {t('cart.remove')}
                </button>
              </figcaption>
            </figure>
          ))}
        </div>

        <label className="flex cursor-pointer items-center gap-2 self-start text-sm underline hover:text-accent">
          {addImage.isPending ? <Spinner className="size-3" /> : '+'} {t('admin.addImage')}
          <input
            type="file"
            accept="image/jpeg,image/png,image/webp"
            className="hidden"
            disabled={addImage.isPending}
            onChange={onPick}
          />
        </label>
      </CardContent>

      {confirmId !== null && (
        <ConfirmDialog
          title={t('admin.deleteImageConfirm')}
          desc=""
          confirmLabel={t('cart.remove')}
          cancelLabel={t('common.cancel')}
          onCancel={() => setConfirmId(null)}
          onConfirm={() => {
            const imgID = confirmId
            setConfirmId(null)
            deleteImage.mutate(imgID, { onError })
          }}
        />
      )}
    </Card>
  )
}
