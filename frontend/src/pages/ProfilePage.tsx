import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { useAuth } from '@/features/auth/AuthContext'
import { useUpdateProfile } from '@/features/auth/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { ApiError } from '@/lib/apiClient'

interface ProfileForm {
  display_name: string
  phone: string
}

export function ProfilePage() {
  const { t } = useLang()
  const { user, setUser, signOut } = useAuth()
  const updateProfile = useUpdateProfile()
  const [saved, setSaved] = useState(false)

  const schema = z.object({
    display_name: z.string().max(100, t('profile.nameTooLong')),
    phone: z.string().refine((v) => v === '' || /^0\d{9}$/.test(v), t('profile.phoneInvalid')),
  })
  const {
    register,
    handleSubmit,
    formState: { errors, isDirty },
  } = useForm<ProfileForm>({
    resolver: zodResolver(schema),
    defaultValues: { display_name: user?.display_name ?? '', phone: user?.phone ?? '' },
  })

  useDocumentTitle(t('profile.docTitle'))

  if (!user) return null // ProtectedRoute การันตีว่ามี user แล้ว

  const onSubmit = handleSubmit((values) => {
    setSaved(false)
    updateProfile.mutate(values, {
      onSuccess: (updated) => {
        setUser(updated)
        setSaved(true)
      },
    })
  })

  const apiMessage = updateProfile.error instanceof ApiError ? updateProfile.error.message : null

  return (
    <div className="mx-auto max-w-2xl px-4 py-12">
      <div className="mb-8 flex items-center justify-between border-b border-border pb-6">
        <div>
          <h1 className="font-display text-3xl uppercase">{t('profile.title')}</h1>
          <p className="mt-1 text-sm text-muted-foreground">{user.email}</p>
        </div>
        <Button variant="outline" size="sm" onClick={() => void signOut()}>
          {t('profile.logout')}
        </Button>
      </div>

      <div className="flex flex-col gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">{t('profile.accountInfo')}</CardTitle>
          </CardHeader>
          <CardContent className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field label={t('profile.email')} value={user.email} />
            <Field
              label={t('profile.role')}
              value={user.role === 'admin' ? t('profile.roleAdmin') : t('profile.roleCustomer')}
            />
            <Field label={t('profile.publicId')} value={user.public_id} mono />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">{t('profile.editProfile')}</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={onSubmit} className="flex flex-col gap-4" noValidate>
              <div className="flex flex-col gap-2">
                <Label htmlFor="display_name">{t('profile.displayName')}</Label>
                <Input
                  id="display_name"
                  placeholder={t('profile.displayNamePh')}
                  {...register('display_name')}
                />
                {errors.display_name && (
                  <p className="text-sm text-destructive">{errors.display_name.message}</p>
                )}
              </div>

              <div className="flex flex-col gap-2">
                <Label htmlFor="phone">{t('profile.phone')}</Label>
                <Input
                  id="phone"
                  inputMode="numeric"
                  placeholder={t('profile.phonePh')}
                  {...register('phone')}
                />
                {errors.phone && <p className="text-sm text-destructive">{errors.phone.message}</p>}
              </div>

              {apiMessage && (
                <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
                  {apiMessage}
                </p>
              )}
              {saved && !updateProfile.isPending && (
                <p className="rounded-md bg-accent/10 px-3 py-2 text-sm text-accent">
                  {t('profile.saved')}
                </p>
              )}

              <Button
                type="submit"
                disabled={updateProfile.isPending || !isDirty}
                className="self-start"
              >
                {updateProfile.isPending && <Spinner />}
                {t('profile.save')}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function Field({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex flex-col gap-1">
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className={mono ? 'break-all font-mono text-sm' : 'text-sm font-medium'}>
        {value || '—'}
      </span>
    </div>
  )
}
