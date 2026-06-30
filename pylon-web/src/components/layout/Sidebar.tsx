import { useNavigate } from 'react-router-dom'

import { useAuth } from '../../hooks/useAuth'
import { getInitial } from '../../utils/format'
import { RoomList } from '../room/RoomList'

type SidebarProps = {
  isOpen: boolean
  onClose: () => void
  onRoomSelect: () => void
}

export function Sidebar({ isOpen, onClose, onRoomSelect }: SidebarProps) {
  const navigate = useNavigate()
  const { logout, user } = useAuth()

  const displayName = userDisplayName(user)
  const username = user?.username.trim() || user?.email.trim() || 'operator'

  function handleLogout() {
    logout()
    onClose()
    navigate('/login', { replace: true })
  }

  return (
    <aside
      aria-label="Primary chat navigation"
      className={`fixed inset-y-0 left-0 z-50 flex w-full flex-col border-r-2 border-[var(--color-ink)] bg-[var(--color-paper)] shadow-[10px_0_0_var(--color-ink)] transition-transform duration-200 md:w-[280px] md:translate-x-0 md:shadow-none ${
        isOpen ? 'translate-x-0' : '-translate-x-full'
      }`}
    >
      <header className="border-b-2 border-[var(--color-ink)] px-4 py-4">
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="font-mono text-[0.65rem] uppercase tracking-[0.26em] text-[var(--color-muted)]">
              Pylon Web
            </p>
            <h1 className="mt-1 text-3xl font-black uppercase leading-none tracking-[-0.06em]">
              Chat
            </h1>
          </div>

          <button
            aria-label="Close sidebar"
            className="border-2 border-[var(--color-ink)] px-3 py-2 font-mono text-xs uppercase tracking-[0.16em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)] md:hidden"
            type="button"
            onClick={onClose}
          >
            X
          </button>
        </div>
      </header>

      <div className="min-h-0 flex-1 overflow-y-auto">
        <RoomList onRoomSelect={onRoomSelect} />
      </div>

      <footer className="border-t-2 border-[var(--color-ink)] px-4 py-4">
        <div className="grid grid-cols-[2.75rem_1fr] gap-3">
          <div className="flex size-11 items-center justify-center overflow-hidden border-2 border-[var(--color-ink)] font-mono text-sm font-black uppercase">
            {user?.avatarUrl ? (
              <img alt="" className="size-full object-cover" src={user.avatarUrl} />
            ) : (
              getInitial(displayName)
            )}
          </div>

          <div className="min-w-0">
            <p className="truncate text-sm font-black uppercase tracking-[-0.03em]">
              {displayName}
            </p>
            <p className="mt-1 truncate font-mono text-[0.65rem] uppercase tracking-[0.16em] text-[var(--color-muted)]">
              @{username}
            </p>
          </div>
        </div>

        <button
          className="mt-4 inline-flex w-full items-center justify-between border-2 border-[var(--color-accent)] px-4 py-3 font-mono text-xs uppercase tracking-[0.18em] text-[var(--color-accent)] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
          type="button"
          onClick={handleLogout}
        >
          Logout
          <span>!</span>
        </button>
      </footer>
    </aside>
  )
}

function userDisplayName(user: { displayName: string; username: string; email: string } | null) {
  if (!user) {
    return 'Operator'
  }

  return user.displayName.trim() || user.username.trim() || user.email.trim() || 'Operator'
}
