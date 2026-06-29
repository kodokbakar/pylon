import { createContext, useContext } from 'react'

import type { UseWebSocketResult } from '../hooks/useWebSocket'

export type WebSocketContextValue = UseWebSocketResult

export const WebSocketContext = createContext<WebSocketContextValue | null>(null)

export function useWebSocketContext() {
  const context = useContext(WebSocketContext)

  if (!context) {
    throw new Error('useWebSocketContext must be used within WebSocketProvider')
  }

  return context
}
