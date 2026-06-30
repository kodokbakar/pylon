import type { RoomListItemData } from '../../hooks/useRooms'
import { StatusIndicator } from '../presence/StatusIndicator'

type RoomItemProps = {
  room: RoomListItemData
  isActive: boolean
  onlineCount: number
  onClick: () => void
}

export function RoomItem({ room, isActive, onlineCount, onClick }: RoomItemProps) {
  return (
    <button
      aria-current={isActive ? 'page' : undefined}
      className={`group grid w-full grid-cols-[3rem_1fr_auto] gap-3 border-b border-[var(--color-line)] px-3 py-4 text-left transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-[var(--color-accent)] ${
        isActive
          ? 'bg-[var(--color-ink)] text-[var(--color-paper)]'
          : 'hover:bg-[var(--color-grid)]'
      }`}
      type="button"
      onClick={onClick}
    >
      <span
        aria-hidden="true"
        className={`flex size-12 items-center justify-center border-2 font-mono text-sm font-black uppercase ${
          isActive
            ? 'border-[var(--color-paper)] text-[var(--color-paper)]'
            : 'border-[var(--color-ink)] text-[var(--color-ink)]'
        }`}
      >
        {room.initial}
      </span>

      <span className="min-w-0">
        <span className="block truncate text-base font-black uppercase leading-tight tracking-[-0.03em]">
          {room.name}
        </span>

        <span
          className={`mt-1 block truncate text-sm leading-5 ${
            isActive ? 'text-[var(--color-line)]' : 'text-[var(--color-muted)]'
          }`}
        >
          {room.lastMessagePreview ?? 'No recent message'}
        </span>
      </span>

      <span className="flex flex-col items-end gap-2">
        <span
          className={`font-mono text-[0.65rem] uppercase tracking-[0.16em] ${
            isActive ? 'text-[var(--color-line)]' : 'text-[var(--color-muted)]'
          }`}
        >
          {formatRoomTimestamp(room.updatedAt)}
        </span>

        <span
          className={`inline-flex items-center gap-1.5 font-mono text-[0.65rem] uppercase tracking-[0.14em] ${
            isActive ? 'text-[var(--color-line)]' : 'text-[var(--color-muted)]'
          }`}
        >
          <StatusIndicator
            label={`${onlineCount} online in ${room.name}`}
            status={onlineCount > 0 ? 'online' : 'offline'}
          />
          {onlineCount}
        </span>

        {room.unreadCount > 0 ? (
          <span
            aria-label={`${room.unreadCount} unread messages`}
            className={`flex min-w-6 items-center justify-center border px-2 py-1 font-mono text-[0.65rem] font-bold ${
              isActive
                ? 'border-[var(--color-paper)] text-[var(--color-paper)]'
                : 'border-[var(--color-accent)] text-[var(--color-accent)]'
            }`}
          >
            {room.unreadCount}
          </span>
        ) : null}
      </span>
    </button>
  )
}

function formatRoomTimestamp(value: string | null) {
  if (!value) {
    return '—'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '—'
  }

  return new Intl.DateTimeFormat('en', {
    month: 'short',
    day: '2-digit',
  }).format(date)
}
