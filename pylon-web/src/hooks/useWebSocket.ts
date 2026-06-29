import { useCallback, useEffect, useRef, useState } from 'react'

import { getWebSocketUrl } from '../api/config'
import {
  PylonWebSocketClient,
  type WebSocketConnectionState,
  type WebSocketMessage,
  type WebSocketMessageHandler,
} from '../lib/ws'

type UseWebSocketOptions = {
  token: string | null
  enabled: boolean
}

export type UseWebSocketResult = {
  state: WebSocketConnectionState
  error: string | null
  reconnectAttempt: number
  maxReconnectAttempts: number
  lastMessage: WebSocketMessage | null
  isConnected: boolean
  send: (message: WebSocketMessage) => boolean
  subscribe: (type: string, handler: WebSocketMessageHandler) => () => void
  connect: () => void
  disconnect: () => void
}

const maxReconnectAttempts = 10

export function useWebSocket({ token, enabled }: UseWebSocketOptions): UseWebSocketResult {
  const clientRef = useRef<PylonWebSocketClient | null>(null)

  const [state, setState] = useState<WebSocketConnectionState>('disconnected')
  const [error, setError] = useState<string | null>(null)
  const [reconnectAttempt, setReconnectAttempt] = useState(0)
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null)

  useEffect(() => {
    if (!enabled || !token) {
      clientRef.current?.disconnect()
      clientRef.current = null
      return
    }

    const client = new PylonWebSocketClient({
      url: getWebSocketUrl(token),
      maxReconnectAttempts,
      heartbeatIntervalMs: 30_000,
      pongTimeoutMs: 10_000,
      reconnectBaseDelayMs: 1_000,
      reconnectMaxDelayMs: 30_000,
      onStateChange: (snapshot) => {
        setState(snapshot.state)
        setReconnectAttempt(snapshot.reconnectAttempt)
        setError(snapshot.error)

        if (snapshot.state === 'disconnected') {
          setLastMessage(null)
        }
      },
      onMessage: (message) => {
        setLastMessage(message)
      },
      onError: (nextError) => {
        setError(nextError.message)
      },
    })

    clientRef.current = client
    client.connect()

    return () => {
      client.disconnect()

      if (clientRef.current === client) {
        clientRef.current = null
      }
    }
  }, [enabled, token])

  const send = useCallback((message: WebSocketMessage) => {
    return clientRef.current?.send(message) ?? false
  }, [])

  const subscribe = useCallback((type: string, handler: WebSocketMessageHandler) => {
    return clientRef.current?.subscribe(type, handler) ?? (() => undefined)
  }, [])

  const connect = useCallback(() => {
    clientRef.current?.connect()
  }, [])

  const disconnect = useCallback(() => {
    clientRef.current?.disconnect()
  }, [])

  return {
    state,
    error,
    reconnectAttempt,
    maxReconnectAttempts,
    lastMessage,
    isConnected: state === 'connected',
    send,
    subscribe,
    connect,
    disconnect,
  }
}
