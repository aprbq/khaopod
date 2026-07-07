import { Badge } from '@/components/ui/badge'
import { Spinner } from '@/components/ui/spinner'
import { useAdminUsers } from '@/features/admin/hooks'
import { useLang } from '@/i18n/LanguageContext'
import { useDocumentTitle } from '@/hooks/useDocumentTitle'
import { formatDateTime } from '@/lib/format'

export function AdminUsersPage() {
  const { t, lang } = useLang()
  useDocumentTitle(t('admin.users'))
  const { data: users, isLoading } = useAdminUsers()

  if (isLoading) {
    return (
      <div className="grid min-h-[40vh] place-items-center">
        <Spinner className="size-6 text-muted-foreground" />
      </div>
    )
  }

  return (
    <div>
      <h1 className="font-display text-2xl uppercase md:text-3xl">{t('admin.users')}</h1>

      {!users || users.length === 0 ? (
        <p className="mt-8 text-sm text-muted-foreground">{t('admin.noUsers')}</p>
      ) : (
        <div className="mt-6 overflow-x-auto border border-border bg-background">
          <table className="w-full min-w-[44rem] text-left text-sm">
            <thead className="border-b border-border text-xs uppercase tracking-wide text-muted-foreground">
              <tr>
                <th className="px-4 py-3">{t('admin.colUser')}</th>
                <th className="px-4 py-3">{t('admin.colRole')}</th>
                <th className="px-4 py-3">{t('admin.colLastLogin')}</th>
                <th className="px-4 py-3">{t('admin.colJoined')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {users.map((u) => (
                <tr key={u.public_id} className="transition-colors hover:bg-muted/50">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      {u.avatar_url ? (
                        <img
                          src={u.avatar_url}
                          alt=""
                          referrerPolicy="no-referrer"
                          loading="lazy"
                          className="size-8 shrink-0 rounded-full object-cover"
                        />
                      ) : (
                        <span
                          aria-hidden
                          className="grid size-8 shrink-0 place-items-center rounded-full bg-primary text-xs uppercase text-primary-foreground"
                        >
                          {(u.display_name || u.email).charAt(0)}
                        </span>
                      )}
                      <div className="min-w-0">
                        <p className="truncate font-medium">{u.display_name || '—'}</p>
                        <p className="truncate text-xs text-muted-foreground">{u.email}</p>
                      </div>
                      {!u.is_active && <Badge variant="outline">{t('admin.inactiveBadge')}</Badge>}
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    {u.role === 'admin' ? (
                      <Badge variant="accent">{t('profile.roleAdmin')}</Badge>
                    ) : (
                      <span className="text-muted-foreground">{t('admin.roleCustomer')}</span>
                    )}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-muted-foreground">
                    {u.last_login_at ? formatDateTime(u.last_login_at, lang) : t('admin.neverLogin')}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-muted-foreground">
                    {formatDateTime(u.created_at, lang)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
