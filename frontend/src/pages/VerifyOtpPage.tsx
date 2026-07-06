import { useEffect, useState } from 'react'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'
import { AuthShell } from '@/components/AuthShell'
import { OtpInput } from '@/components/OtpInput'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Spinner } from '@/components/ui/spinner'
import { useAuth } from '@/features/auth/AuthContext'
import { useRequestOtp, useVerifyOtp } from '@/features/auth/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { ApiError } from '@/lib/apiClient'

const OTP_LENGTH = 6

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

  const [code, setCode] = useState('')
  const [expiry, setExpiry] = useState(state.expiresIn ?? 300)
  const [cooldown, setCooldown] = useState(60) // rate limit backend: ขอใหม่ได้ทุก 60 วิ

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

  const submit = (value = code) => {
    if (value.length !== OTP_LENGTH || verifyOtp.isPending) return
    verifyOtp.mutate(
      { email, code: value },
      {
        onSuccess: (res) => {
          setSession(res)
          navigate('/account', { replace: true })
        },
      },
    )
  }

  // แก้เลขเมื่อไหร่ ล้าง error เก่าทิ้ง (ให้ช่องหายแดง) แล้ว auto-submit เมื่อครบ
  const onCodeChange = (next: string) => {
    if (verifyOtp.error) verifyOtp.reset()
    setCode(next)
  }

  const resend = () => {
    setCode('')
    verifyOtp.reset()
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
          <form
            onSubmit={(e) => {
              e.preventDefault()
              submit()
            }}
            className="flex flex-col gap-4"
            noValidate
          >
            <OtpInput
              value={code}
              onChange={onCodeChange}
              onComplete={(c) => submit(c)}
              length={OTP_LENGTH}
              disabled={verifyOtp.isPending}
              invalid={!!verifyMessage}
              ariaLabel={t('verify.otp')}
            />

            <p className="text-center text-xs text-muted-foreground">
              {expiry > 0 ? t('verify.expiresIn', { time: mmss(expiry) }) : t('verify.expired')}
            </p>

            {verifyMessage && (
              <p className="rounded-md bg-destructive/10 px-3 py-2 text-center text-sm text-destructive">
                {verifyMessage}
              </p>
            )}

            <Button
              type="submit"
              size="lg"
              disabled={code.length !== OTP_LENGTH || verifyOtp.isPending}
            >
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
