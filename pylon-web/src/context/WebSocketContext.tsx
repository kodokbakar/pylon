import type { PropsWithChildren } from 'react'

import { useAuth } from '../hooks/useAuth'
import { useWebSocket } from '../hooks/useWebSocket'
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
      <WebSocketErrorBanner error={webSocket.error} state={webSocket.state} />
    </WebSocketContext.Provider>
  )
}

function WebSocketErrorBanner({ error, state }: { error: string | null; state: string }) {
  if (state !== 'error' || !error) {
    return null
  }

  return (
    <div
      className="fixed bottom-4 right-4 z-50 max-w-sm border-2 border-[var(--color-accent)] bg-[var(--color-paper)] px-4 py-3 shadow-[6px_6px_0_var(--color-ink)]"
      role="alert"
    >
      <p className="font-mono text-[0.65rem] uppercase tracking-[0.18em] text-[var(--color-accent)]">
        Realtime connection failed
      </p>
      <p className="mt-2 text-sm font-semibold leading-5 text-[var(--color-ink)]">{error}</p>
    </div>
  )
}
