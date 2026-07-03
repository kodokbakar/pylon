import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { UseWebSocketResult } from '../../hooks/useWebSocket'
import type { StoredAuthUser } from '../../utils/authToken'
import { WebSocketProvider } from '../WebSocketContext'

const mockAuth = vi.hoisted(() => ({
  token: 'token-1',
  isAuthenticated: true,
  user: {
    id: 'user-1',
    username: 'operator',
    email: 'operator@example.com',
    displayName: 'Operator',
    avatarUrl: '',
    createdAt: '2026-07-01T00:00:00.000Z',
  } satisfies StoredAuthUser,
  login: vi.fn(),
  logout: vi.fn(),
}))

const mockWebSocket = vi.hoisted(() => ({
  value: createWebSocketValue(),
}))

vi.mock('../../hooks/useAuth', () => ({
  useAuth: () => mockAuth,
}))

vi.mock('../../hooks/useWebSocket', () => ({
  useWebSocket: () => mockWebSocket.value,
}))

describe('WebSocketProvider', () => {
  beforeEach(() => {
    mockWebSocket.value = createWebSocketValue()
  })

  it('renders children when realtime state is healthy', () => {
    render(
      <WebSocketProvider>
        <p>Realtime children</p>
      </WebSocketProvider>,
    )

    expect(screen.getByText('Realtime children')).toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('renders reconnect attempt metadata while reconnecting', () => {
    mockWebSocket.value = createWebSocketValue({
      state: 'reconnecting',
      error: 'WebSocket connection error',
      reconnectAttempt: 2,
      nextReconnectDelayMs: 2_000,
    })

    render(
      <WebSocketProvider>
        <p>Realtime children</p>
      </WebSocketProvider>,
    )

    expect(screen.getByRole('alert')).toHaveTextContent('Realtime reconnecting')
    expect(screen.getByRole('alert')).toHaveTextContent('Attempt 2/10 · next retry in 2s')
  })

  it('calls manual reconnect from the retry button', async () => {
    const user = userEvent.setup()
    const connect = vi.fn()

    mockWebSocket.value = createWebSocketValue({
      state: 'error',
      error: 'WebSocket reconnect attempts exhausted',
      reconnectAttempt: 10,
      connect,
    })

    render(
      <WebSocketProvider>
        <p>Realtime children</p>
      </WebSocketProvider>,
    )

    await user.click(screen.getByRole('button', { name: /retry now/i }))

    expect(connect).toHaveBeenCalledTimes(1)
  })
})

function createWebSocketValue(overrides: Partial<UseWebSocketResult> = {}): UseWebSocketResult {
  return {
    state: 'connected',
    error: null,
    reconnectAttempt: 0,
    maxReconnectAttempts: 10,
    nextReconnectDelayMs: null,
    lastMessage: null,
    isConnected: true,
    send: vi.fn<UseWebSocketResult['send']>(() => true),
    subscribe: vi.fn<UseWebSocketResult['subscribe']>(() => () => undefined),
    connect: vi.fn<UseWebSocketResult['connect']>(),
    disconnect: vi.fn<UseWebSocketResult['disconnect']>(),
    ...overrides,
  }
}
