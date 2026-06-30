import type { RoomPresenceStatus } from '../../hooks/useRoomPresence'
import { formatPresenceStatus } from '../../hooks/useRoomPresence'

type StatusIndicatorProps = {
  status: RoomPresenceStatus
  label?: string
  size?: 'sm' | 'md'
  className?: string
}

export function StatusIndicator({
  status,
  label,
  size = 'sm',
  className = '',
}: StatusIndicatorProps) {
  const sizeClass = size === 'md' ? 'size-3' : 'size-2.5'
  const statusClass = statusToClassName(status)
  const accessibleLabel = label ?? `User is ${formatPresenceStatus(status)}`

  return (
    <span
      aria-label={accessibleLabel}
      className={`inline-flex items-center justify-center ${className}`}
      title={accessibleLabel}
    >
      <span
        className={`${sizeClass} ${statusClass} block border border-[var(--color-ink)]`}
        aria-hidden="true"
      />
    </span>
  )
}

function statusToClassName(status: RoomPresenceStatus) {
  switch (status) {
    case 'online':
      return 'bg-[var(--color-presence-online)]'
    case 'typing':
      return 'relative bg-[var(--color-presence-typing)] before:absolute before:inset-[-0.25rem] before:animate-ping before:bg-[var(--color-presence-typing)] before:opacity-35'
    default:
      return 'bg-[var(--color-presence-offline)]'
  }
}
