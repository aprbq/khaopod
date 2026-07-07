import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { useLang } from '@/i18n/LanguageContext'
import type { AddressInput } from '@/types/address'

interface AddressFormProps {
  pending: boolean
  onSubmit: (input: AddressInput) => void
  onCancel?: () => void
}

// ฟอร์มเพิ่มที่อยู่จัดส่ง (โครงสร้างแบบไทย) — validation ฝั่ง client เป็นแค่ UX, server ตรวจซ้ำเสมอ
export function AddressForm({ pending, onSubmit, onCancel }: AddressFormProps) {
  const { t } = useLang()

  const schema = z.object({
    recipient_name: z.string().min(1, t('addr.recipientRequired')),
    phone: z.string().regex(/^0\d{9}$/, t('profile.phoneInvalid')),
    address_line: z.string().min(1, t('addr.required')),
    subdistrict: z.string().min(1, t('addr.required')),
    district: z.string().min(1, t('addr.required')),
    province: z.string().min(1, t('addr.required')),
    postal_code: z.string().regex(/^\d{5}$/, t('addr.postalInvalid')),
  })
  type Form = z.infer<typeof schema>

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<Form>({
    resolver: zodResolver(schema),
    defaultValues: {
      recipient_name: '',
      phone: '',
      address_line: '',
      subdistrict: '',
      district: '',
      province: '',
      postal_code: '',
    },
  })

  const field = (
    name: keyof Form,
    label: string,
    opts?: { className?: string; inputMode?: 'numeric' },
  ) => (
    <div className={`flex flex-col gap-1.5 ${opts?.className ?? ''}`}>
      <Label htmlFor={`addr-${name}`}>{label}</Label>
      <Input id={`addr-${name}`} inputMode={opts?.inputMode} {...register(name)} />
      {errors[name] && <p className="text-sm text-destructive">{errors[name]?.message}</p>}
    </div>
  )

  return (
    <form
      onSubmit={handleSubmit((v) => onSubmit(v))}
      className="grid grid-cols-1 gap-3 sm:grid-cols-2"
      noValidate
    >
      {field('recipient_name', t('addr.recipient'))}
      {field('phone', t('profile.phone'), { inputMode: 'numeric' })}
      {field('address_line', t('addr.line'), { className: 'sm:col-span-2' })}
      {field('subdistrict', t('addr.subdistrict'))}
      {field('district', t('addr.district'))}
      {field('province', t('addr.province'))}
      {field('postal_code', t('addr.postalCode'), { inputMode: 'numeric' })}

      <div className="flex gap-3 sm:col-span-2">
        <Button type="submit" disabled={pending}>
          {pending && <Spinner />}
          {t('addr.save')}
        </Button>
        {onCancel && (
          <Button type="button" variant="outline" disabled={pending} onClick={onCancel}>
            {t('common.cancel')}
          </Button>
        )}
      </div>
    </form>
  )
}
