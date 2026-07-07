import { zodResolver } from '@hookform/resolvers/zod'
import { useRef, useState } from 'react'
import type { ChangeEvent } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { useAuth } from '@/features/auth/AuthContext'
import { useUpdateProfile, useUploadAvatar } from '@/features/auth/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { ApiError } from '@/lib/apiClient'
import { formatDate } from '@/lib/format'

const MAX_AVATAR_BYTES = 2 * 1024 * 1024 // ต้องตรงกับเพดานฝั่ง backend (docs/rest_api.md §3.3)

interface ProfileForm {
  display_name: string
  phone: string
}

export function ProfilePage() {
  const { t, lang } = useLang()
  const { user, setUser, signOut } = useAuth()
  const updateProfile = useUpdateProfile()
  const uploadAvatar = useUploadAvatar()
  const fileRef = useRef<HTMLInputElement>(null)
  const [avatarErr, setAvatarErr] = useState('')
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

  const onPickAvatar = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = '' // เคลียร์ค่าให้เลือกไฟล์เดิมซ้ำแล้ว change ยิงอีกครั้งได้
    if (!file) return
    if (file.size > MAX_AVATAR_BYTES) {
      setAvatarErr(t('profile.avatarTooBig'))
      return
    }
    setAvatarErr('')
    uploadAvatar.mutate(file, {
      onSuccess: (updated) => setUser(updated),
      onError: (err) => setAvatarErr(err instanceof ApiError ? err.message : t('common.error')),
    })
  }

  // ชื่อที่โชว์บน header — ถ้ายังไม่ตั้งชื่อ ใช้ส่วนหน้า @ ของอีเมลแทน ไม่ปล่อยว่าง
  const displayName = user.display_name || user.email.split('@')[0]

  return (
    <div className="mx-auto max-w-2xl px-4 py-12">
      <div className="mb-8 flex flex-wrap items-center gap-4 border-b border-border pb-6">
        {user.avatar_url ? (
          // no-referrer: รูปจาก googleusercontent จะ 403 ถ้าแนบ referrer ข้ามโดเมน
          <img
            src={user.avatar_url}
            alt=""
            referrerPolicy="no-referrer"
            className="size-16 shrink-0 rounded-full object-cover"
          />
        ) : (
          <div
            aria-hidden
            className="grid size-16 shrink-0 place-items-center rounded-full bg-primary font-display text-2xl uppercase text-primary-foreground"
          >
            {displayName.charAt(0)}
          </div>
        )}
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <h1 className="truncate font-display text-2xl uppercase md:text-3xl">{displayName}</h1>
            {user.role === 'admin' && <Badge variant="accent">{t('profile.roleAdmin')}</Badge>}
          </div>
          <p className="mt-1 truncate text-sm text-muted-foreground">{user.email}</p>
          <p className="mt-1 text-xs text-muted-foreground">
            {t('profile.joined', { date: formatDate(user.created_at, lang) })}
            {' · '}
            {t('profile.lastUpdated', { date: formatDate(user.updated_at, lang) })}
          </p>
          <input
            ref={fileRef}
            type="file"
            accept="image/jpeg,image/png,image/webp"
            className="hidden"
            onChange={onPickAvatar}
          />
          <button
            type="button"
            className="mt-2 inline-flex items-center gap-1.5 text-sm underline hover:text-accent disabled:opacity-40"
            disabled={uploadAvatar.isPending}
            onClick={() => fileRef.current?.click()}
          >
            {uploadAvatar.isPending && <Spinner className="size-3" />}
            {t('profile.changeAvatar')}
          </button>
          {avatarErr && <p className="mt-1 text-xs text-destructive">{avatarErr}</p>}
        </div>
        <Button variant="outline" size="sm" onClick={() => void signOut()}>
          {t('profile.logout')}
        </Button>
      </div>

      <div className="flex flex-col gap-6">
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
