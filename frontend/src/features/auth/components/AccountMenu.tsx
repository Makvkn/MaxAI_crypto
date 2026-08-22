import { useNavigate } from 'react-router-dom'
import { Badge } from '@/components/ui/Badge'
import { Menu } from '@/components/ui/Menu'
import { User } from '@/components/ui/Icon'
import { useSession } from '../sessionContext'

/**
 * Account menu.
 *
 * A guest is a real account with real data, so the menu offers an upgrade
 * rather than a sign-up that would start over.
 */
export function AccountMenu() {
  const navigate = useNavigate()
  const { user, isGuest, signOut } = useSession()

  return (
    <Menu
      label="Account"
      align="end"
      items={[
        ...(isGuest
          ? [
              {
                id: 'upgrade',
                label: 'Save my analysis',
                description: 'Keep this data with an email or Google account',
                onSelect: () => navigate('/sign-in?upgrade=1'),
              },
            ]
          : []),
        {
          id: 'sign-out',
          label: 'Sign out',
          onSelect: () => {
            void signOut().then(() => navigate('/'))
          },
        },
      ]}
      trigger={(triggerProps) => (
        <button
          type="button"
          {...triggerProps}
          className="flex items-center gap-2 rounded-lg border border-line bg-surface px-2.5 py-2 text-[12px] text-fg-muted transition-colors hover:border-line-strong hover:text-fg"
        >
          <User className="size-4" />
          <span className="hidden sm:inline">
            {user?.email ?? (isGuest ? 'Guest' : 'Account')}
          </span>
          {isGuest ? (
            <Badge tone="caution" className="hidden sm:inline-flex">
              Unsaved
            </Badge>
          ) : null}
        </button>
      )}
    />
  )
}
