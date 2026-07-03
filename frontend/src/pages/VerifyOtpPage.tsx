import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'
import { z } from 'zod'
import { AuthShell } from '@/components/AuthShell'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { useAuth } from '@/features/auth/AuthContext'
import { useRequestOtp, useVerifyOtp } from '@/features/auth/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { ApiError } from '@/lib/apiClient'

interface VerifyForm {
  code: string
}

interface LocationState {
  email?: string
  expiresIn?: number
}

function mmss(total: number): string {
  const m = Math.floor(total / 60)
  const s = total % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

export function VerifyOtpPage() {
  const { t } = useLang()
  useDocumentTitle(t('verify.docTitle'))
  const navigate = useNavigate()
  const location = useLocation()
  const { setSession } = useAuth()
  const state = (location.state ?? {}) as LocationState

  const verifyOtp = useVerifyOtp()
  const requestOtp = useRequestOtp()

  const [expiry, setExpiry] = useState(state.expiresIn ?? 300)
  const [cooldown, setCooldown] = useState(60) // rate limit backend: ขอใหม่ได้ทุก 60 วิ

  const schema = z.object({ code: z.string().regex(/^\d{6}$/, t('verify.otpInvalid')) })
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<VerifyForm>({ resolver: zodResolver(schema), defaultValues: { code: '' } })

  useEffect(() => {
    const id = setInterval(() => {
      setExpiry((s) => (s > 0 ? s - 1 : 0))
      setCooldown((s) => (s > 0 ? s - 1 : 0))
    }, 1000)
    return () => clearInterval(id)
  }, [])

  // ไม่มีอีเมลใน state (เข้าหน้านี้ตรง ๆ) → กลับไปเริ่มที่ login
  if (!state.email) {
    return <Navigate to="/login" replace />
  }
  const email = state.email

  const onSubmit = handleSubmit((values) => {
    verifyOtp.mutate(
      { email, code: values.code },
      {
        onSuccess: (res) => {
          setSession(res)
          navigate('/account', { replace: true })
        },
      },
    )
  })

  const resend = () => {
    requestOtp.mutate(email, {
      onSuccess: (res) => {
        setExpiry(res.expires_in)
        setCooldown(60)
      },
    })
  }

  const verifyMessage = verifyOtp.error instanceof ApiError ? verifyOtp.error.message : null
  const resendMessage = requestOtp.error instanceof ApiError ? requestOtp.error.message : null

  return (
    <AuthShell>
      <Card>
        <CardHeader>
          <CardTitle>{t('verify.title')}</CardTitle>
          <CardDescription>
            {t('verify.sentTo').split('{email}')[0]}
            <span className="font-medium text-foreground">{email}</span>
            {t('verify.sentTo').split('{email}')[1]}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} className="flex flex-col gap-4" noValidate>
            <div className="flex flex-col gap-2">
              <Label htmlFor="code">{t('verify.otp')}</Label>
              <Input
                id="code"
                inputMode="numeric"
                autoComplete="one-time-code"
                maxLength={6}
                placeholder="000000"
                className="text-center text-lg tracking-[0.5em]"
                aria-invalid={!!errors.code}
                {...register('code')}
              />
              {errors.code && <p className="text-sm text-destructive">{errors.code.message}</p>}
              <p className="text-xs text-muted-foreground">
                {expiry > 0 ? t('verify.expiresIn', { time: mmss(expiry) }) : t('verify.expired')}
              </p>
            </div>

            {verifyMessage && (
              <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {verifyMessage}
              </p>
            )}

            <Button type="submit" size="lg" disabled={verifyOtp.isPending}>
              {verifyOtp.isPending && <Spinner />}
              {t('verify.submit')}
            </Button>
          </form>

          <div className="mt-4 flex flex-col items-center gap-1">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={resend}
              disabled={cooldown > 0 || requestOtp.isPending}
            >
              {requestOtp.isPending && <Spinner />}
              {cooldown > 0 ? t('verify.resendIn', { sec: cooldown }) : t('verify.resend')}
            </Button>
            {resendMessage && <p className="text-xs text-destructive">{resendMessage}</p>}
          </div>
        </CardContent>
      </Card>
    </AuthShell>
  )
}
