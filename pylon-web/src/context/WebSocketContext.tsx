import type { PropsWithChildren } from 'react'

import { useAuth } from '../hooks/useAuth'
import { useWebSocket, type UseWebSocketResult } from '../hooks/useWebSocket'
import { ConnectionStatus } from '../components/ui/ConnectionStatus'
import { WebSocketContext } from './webSocketContext'

export function WebSocketProvider({ children }: PropsWithChildren) {
  const { token, isAuthenticated } = useAuth()

  const webSocket = useWebSocket({
    token,
    enabled: isAuthenticated && Boolean(token),
  })

  return (
    <WebSocketContext.Provider value={webSocket}>
      {children}
      <WebSocketErrorBanner webSocket={webSocket} />
    </WebSocketContext.Provider>
  )
}

function WebSocketErrorBanner({ webSocket }: { webSocket: UseWebSocketResult }) {
  const shouldShow =
    webSocket.state === 'reconnecting' ||
    webSocket.state === 'error' ||
    (webSocket.state === 'disconnected' && Boolean(webSocket.error))

  if (!shouldShow) {
    return null
  }

  const retryDelay = formatDelay(webSocket.nextReconnectDelayMs)
  const isRetrying = webSocket.state === 'reconnecting'

  return (
    <div
      className="fixed bottom-4 right-4 z-50 max-w-sm border-2 border-[var(--color-accent)] bg-[var(--color-paper)] px-4 py-3 shadow-[6px_6px_0_var(--color-ink)]"
      role="alert"
    >
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="font-mono text-[0.65rem] uppercase tracking-[0.18em] text-[var(--color-accent)]">
            {isRetrying ? 'Realtime reconnecting' : 'Realtime connection failed'}
          </p>

          <p className="mt-2 text-sm font-semibold leading-5 text-[var(--color-ink)]">
            {webSocket.error ??
              `Reconnect attempt ${webSocket.reconnectAttempt}/${webSocket.maxReconnectAttempts}`}
          </p>
        </div>

        <ConnectionStatus
          compact
          maxReconnectAttempts={webSocket.maxReconnectAttempts}
          nextReconnectDelayMs={webSocket.nextReconnectDelayMs}
          reconnectAttempt={webSocket.reconnectAttempt}
          state={webSocket.state}
        />
      </div>

      <p className="mt-3 font-mono text-[0.65rem] uppercase tracking-[0.14em] text-[var(--color-muted)]">
        Attempt {webSocket.reconnectAttempt}/{webSocket.maxReconnectAttempts}
        {retryDelay ? ` · next retry in ${retryDelay}` : ''}
      </p>

      <button
        className="mt-4 inline-flex w-full items-center justify-between border-2 border-[var(--color-ink)] px-4 py-3 font-mono text-xs uppercase tracking-[0.18em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
        type="button"
        onClick={webSocket.connect}
      >
        Retry now
        <span>↻</span>
      </button>
    </div>
  )
}

function formatDelay(delayMs: number | null) {
  if (!delayMs || delayMs <= 0) {
    return ''
  }

  return `${Math.ceil(delayMs / 1_000)}s`
}
