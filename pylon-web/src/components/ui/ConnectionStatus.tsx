import type { WebSocketConnectionState } from '../../lib/ws'

type ConnectionStatusProps = {
  state: WebSocketConnectionState
  reconnectAttempt: number
  maxReconnectAttempts: number
  nextReconnectDelayMs?: number | null
  compact?: boolean
  className?: string
}

export function ConnectionStatus({
  state,
  reconnectAttempt,
  maxReconnectAttempts,
  nextReconnectDelayMs = null,
  compact = false,
  className = '',
}: ConnectionStatusProps) {
  const view = getConnectionStatusView(
    state,
    reconnectAttempt,
    maxReconnectAttempts,
    nextReconnectDelayMs,
  )

  return (
    <div
      aria-label={view.ariaLabel}
      className={[
        'inline-flex items-center gap-2 border border-[var(--color-ink)] bg-[var(--color-paper)] font-mono text-[0.65rem] uppercase tracking-[0.16em]',
        compact ? 'px-2 py-1' : 'px-3 py-2',
        className,
      ]
        .filter(Boolean)
        .join(' ')}
      title={view.ariaLabel}
    >
      <span
        aria-hidden="true"
        className={['size-2.5 border border-[var(--color-ink)]', view.dotClassName].join(' ')}
      />

      <span>{compact ? view.shortLabel : view.label}</span>
    </div>
  )
}

function getConnectionStatusView(
  state: WebSocketConnectionState,
  reconnectAttempt: number,
  maxReconnectAttempts: number,
  nextReconnectDelayMs: number | null,
) {
  switch (state) {
    case 'connected':
      return {
        label: 'Connected',
        shortLabel: 'Live',
        ariaLabel: 'Realtime connection connected',
        dotClassName: 'bg-[var(--color-presence-online)]',
      }

    case 'connecting':
      return {
        label: 'Connecting',
        shortLabel: 'Init',
        ariaLabel: 'Realtime connection connecting',
        dotClassName: 'bg-[var(--color-muted)]',
      }

    case 'reconnecting': {
      const retryDelay = formatDelay(nextReconnectDelayMs)
      const label = retryDelay
        ? `Reconnecting ${reconnectAttempt}/${maxReconnectAttempts} · ${retryDelay}`
        : `Reconnecting ${reconnectAttempt}/${maxReconnectAttempts}`

      return {
        label,
        shortLabel: `Retry ${reconnectAttempt}`,
        ariaLabel: `Realtime connection reconnecting, attempt ${reconnectAttempt} of ${maxReconnectAttempts}${
          retryDelay ? `, next retry in ${retryDelay}` : ''
        }`,
        dotClassName: 'bg-[var(--color-presence-typing)]',
      }
    }

    case 'error':
      return {
        label: `Disconnected ${reconnectAttempt}/${maxReconnectAttempts}`,
        shortLabel: 'Down',
        ariaLabel: `Realtime connection disconnected after ${reconnectAttempt} of ${maxReconnectAttempts} reconnect attempts`,
        dotClassName: 'bg-[var(--color-accent)]',
      }

    case 'disconnected':
    default:
      return {
        label: 'Disconnected',
        shortLabel: 'Off',
        ariaLabel: 'Realtime connection disconnected',
        dotClassName: 'bg-[var(--color-accent)]',
      }
  }
}

function formatDelay(delayMs: number | null) {
  if (!delayMs || delayMs <= 0) {
    return ''
  }

  return `${Math.ceil(delayMs / 1_000)}s`
}
