import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { z } from 'zod'
import { AuthShell } from '@/components/AuthShell'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { useRequestOtp } from '@/features/auth/hooks'
import { ApiError } from '@/lib/apiClient'

const schema = z.object({
  email: z.email('อีเมลไม่ถูกต้อง'),
})
type LoginForm = z.infer<typeof schema>

export function LoginPage() {
  const navigate = useNavigate()
  const requestOtp = useRequestOtp()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginForm>({ resolver: zodResolver(schema), defaultValues: { email: '' } })

  const onSubmit = handleSubmit((values) => {
    requestOtp.mutate(values.email, {
      onSuccess: (res) => navigate('/verify', { state: { email: res.email, expiresIn: res.expires_in } }),
    })
  })

  const apiMessage = requestOtp.error instanceof ApiError ? requestOtp.error.message : null

  return (
    <AuthShell>
      <Card>
        <CardHeader>
          <CardTitle>เข้าสู่ระบบ</CardTitle>
          <CardDescription>กรอกอีเมลเพื่อรับรหัส OTP — ไม่ต้องใช้รหัสผ่าน</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} className="flex flex-col gap-4" noValidate>
            <div className="flex flex-col gap-2">
              <Label htmlFor="email">อีเมล</Label>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                placeholder="you@example.com"
                aria-invalid={!!errors.email}
                {...register('email')}
              />
              {errors.email && <p className="text-sm text-destructive">{errors.email.message}</p>}
            </div>

            {apiMessage && (
              <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {apiMessage}
              </p>
            )}

            <Button type="submit" size="lg" disabled={requestOtp.isPending}>
              {requestOtp.isPending && <Spinner />}
              ขอรหัส OTP
            </Button>
          </form>
        </CardContent>
      </Card>
    </AuthShell>
  )
}
