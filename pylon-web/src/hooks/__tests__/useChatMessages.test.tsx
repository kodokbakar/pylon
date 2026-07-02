import { QueryClientProvider } from '@tanstack/react-query'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { AuthProvider } from '../../context/AuthContext'
import { WebSocketContext, type WebSocketContextValue } from '../../context/webSocketContext'
import { createTestQueryClient } from '../../test/render'
import type { ChatHistoryMessage } from '../../api/chat'
import type { WebSocketMessage, WebSocketMessageHandler } from '../../lib/ws'

const mockGetMessages = vi.hoisted(() => vi.fn())

vi.mock('../../api/chat', () => ({
  ChatService: {
    getMessages: mockGetMessages,
  },
}))

import { useChatMessages } from '../useChatMessages'

describe('useChatMessages', () => {
  beforeEach(() => {
    mockGetMessages.mockReset()
  })

  it('loads history messages for the current room', async () => {
    mockGetMessages.mockResolvedValueOnce({
      messages: [
        createHistoryMessage({
          id: 'message-1',
          content: 'History hello',
        }),
      ],
      hasMore: true,
    })

    renderUseChatMessages()

    expect(await screen.findByText('History hello')).toBeInTheDocument()
    expect(screen.getByTestId('has-more')).toHaveTextContent('true')
    expect(screen.getByText('persisted')).toBeInTheDocument()

    expect(mockGetMessages).toHaveBeenCalledTimes(1)
    expect(mockGetMessages).toHaveBeenCalledWith({
      roomId: 'room-1',
      limit: 50,
    })
  })

  it('adds an optimistic sending message when sendMessage succeeds', async () => {
    const websocket = createMockWebSocket()
    mockGetMessages.mockResolvedValueOnce({
      messages: [],
      hasMore: false,
    })

    renderUseChatMessages({ websocket })

    await waitFor(() => {
      expect(mockGetMessages).toHaveBeenCalledTimes(1)
    })

    fireEvent.click(screen.getByRole('button', { name: 'Send optimistic message' }))

    expect(await screen.findByText('Optimistic hello')).toBeInTheDocument()
    expect(screen.getByText('sending')).toBeInTheDocument()
    expect(screen.getByText('optimistic')).toBeInTheDocument()

    expect(websocket.value.send).toHaveBeenCalledWith({
      type: 'chat.message',
      roomId: 'room-1',
      content: 'Optimistic hello',
      msgType: 'text',
    })
  })

  it('merges real-time messages from the WebSocket subscription', async () => {
    const websocket = createMockWebSocket()
    mockGetMessages.mockResolvedValueOnce({
      messages: [],
      hasMore: false,
    })

    renderUseChatMessages({ websocket })

    await waitFor(() => {
      expect(websocket.value.subscribe).toHaveBeenCalledWith('message', expect.any(Function))
    })

    act(() => {
      websocket.emit('message', {
        type: 'message',
        room_id: 'room-1',
        message_id: 'realtime-1',
        sender_id: 'user-2',
        sender_username: 'teammate',
        sender_name: 'Teammate',
        content: 'Realtime hello',
        msg_type: 'text',
        created_at: '2026-07-01T00:00:01.000Z',
      })
    })

    expect(await screen.findByText('Realtime hello')).toBeInTheDocument()
    expect(screen.getByText('Teammate')).toBeInTheDocument()
    expect(screen.getByText('persisted')).toBeInTheDocument()
  })

  it('deduplicates a matching optimistic message when the server echo arrives', async () => {
    const websocket = createMockWebSocket()
    mockGetMessages.mockResolvedValueOnce({
      messages: [],
      hasMore: false,
    })

    renderUseChatMessages({ websocket })

    await waitFor(() => {
      expect(mockGetMessages).toHaveBeenCalledTimes(1)
    })

    fireEvent.click(screen.getByRole('button', { name: 'Send optimistic message' }))

    expect(await screen.findByText('Optimistic hello')).toBeInTheDocument()
    expect(screen.getAllByTestId('chat-message')).toHaveLength(1)

    act(() => {
      websocket.emit('message', {
        type: 'message',
        room_id: 'room-1',
        message_id: 'server-message-1',
        sender_id: 'user-1',
        sender_username: 'operator',
        sender_name: 'operator',
        content: 'Optimistic hello',
        msg_type: 'text',
        created_at: new Date().toISOString(),
      })
    })

    await waitFor(() => {
      expect(screen.getAllByText('Optimistic hello')).toHaveLength(1)
    })

    expect(screen.getAllByTestId('chat-message')).toHaveLength(1)
    expect(screen.getByText('server-message-1')).toBeInTheDocument()
    expect(screen.getByText('persisted')).toBeInTheDocument()
  })

  it('exposes history loading errors', async () => {
    mockGetMessages.mockRejectedValueOnce(new Error('History failed'))

    renderUseChatMessages()

    await waitFor(() => {
      expect(screen.getByTestId('error-message')).toHaveTextContent('History failed')
    })
  })

  it('loads older messages before the oldest persisted message', async () => {
    mockGetMessages
      .mockResolvedValueOnce({
        messages: [
          createHistoryMessage({
            id: 'message-latest',
            content: 'Latest message',
            createdAt: '2026-07-01T00:05:00.000Z',
          }),
        ],
        hasMore: true,
      })
      .mockResolvedValueOnce({
        messages: [
          createHistoryMessage({
            id: 'message-older',
            content: 'Older message',
            createdAt: '2026-07-01T00:00:00.000Z',
          }),
        ],
        hasMore: false,
      })

    renderUseChatMessages()

    expect(await screen.findByText('Latest message')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Load older messages' }))

    expect(await screen.findByText('Older message')).toBeInTheDocument()

    expect(mockGetMessages).toHaveBeenNthCalledWith(2, {
      roomId: 'room-1',
      limit: 50,
      beforeId: 'message-latest',
    })

    expect(screen.getAllByTestId('chat-message')).toHaveLength(2)
    expect(screen.getByTestId('has-more')).toHaveTextContent('false')
  })
})

function renderUseChatMessages({
  roomId = 'room-1',
  websocket = createMockWebSocket(),
}: {
  roomId?: string
  websocket?: MockWebSocket
} = {}) {
  storeAuthSession()

  const queryClient = createTestQueryClient()

  return {
    websocket,
    queryClient,
    ...render(
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <WebSocketContext.Provider value={websocket.value}>
            <ChatMessagesProbe roomId={roomId} />
          </WebSocketContext.Provider>
        </AuthProvider>
      </QueryClientProvider>,
    ),
  }
}

function ChatMessagesProbe({ roomId }: { roomId: string | undefined }) {
  const chat = useChatMessages(roomId)

  return (
    <section>
      <p data-testid="is-loading">{String(chat.isLoading)}</p>
      <p data-testid="has-more">{String(chat.hasMore)}</p>
      <p data-testid="error-message">{chat.errorMessage ?? ''}</p>
      <p data-testid="send-error">{chat.sendError ?? ''}</p>

      <button type="button" onClick={() => chat.sendMessage('Optimistic hello')}>
        Send optimistic message
      </button>

      <button type="button" onClick={() => void chat.loadOlder()}>
        Load older messages
      </button>

      <ul aria-label="chat messages">
        {chat.messages.map((message) => (
          <li data-testid="chat-message" key={message.id}>
            <span>{message.id}</span>
            <span>{message.content}</span>
            <span>{message.senderName}</span>
            <span>{message.status}</span>
            <span>{message.optimistic ? 'optimistic' : 'persisted'}</span>
          </li>
        ))}
      </ul>
    </section>
  )
}

type MockWebSocket = {
  value: WebSocketContextValue
  emit: (type: string, message: WebSocketMessage) => void
}

function createMockWebSocket(): MockWebSocket {
  const handlers = new Map<string, Set<WebSocketMessageHandler>>()

  const value: WebSocketContextValue = {
    state: 'connected',
    error: null,
    reconnectAttempt: 0,
    maxReconnectAttempts: 10,
    lastMessage: null,
    isConnected: true,
    send: vi.fn(() => true),
    subscribe: vi.fn((type, handler) => {
      const currentHandlers = handlers.get(type) ?? new Set<WebSocketMessageHandler>()
      currentHandlers.add(handler)
      handlers.set(type, currentHandlers)

      return () => {
        currentHandlers.delete(handler)
      }
    }),
    connect: vi.fn(),
    disconnect: vi.fn(),
  }

  return {
    value,
    emit(type, message) {
      handlers.get(type)?.forEach((handler) => handler(message))
    },
  }
}

function createHistoryMessage(overrides: Partial<ChatHistoryMessage> = {}): ChatHistoryMessage {
  return {
    id: 'message-1',
    roomId: 'room-1',
    senderId: 'user-2',
    senderUsername: 'teammate',
    senderDisplayName: 'Teammate',
    senderAvatarUrl: '',
    content: 'History message',
    msgType: 'text',
    createdAt: '2026-07-01T00:00:00.000Z',
    ...overrides,
  }
}

function storeAuthSession() {
  window.localStorage.setItem('auth_token', 'access-token')
  window.localStorage.setItem('refresh_token', 'refresh-token')
  window.localStorage.setItem(
    'auth_user',
    JSON.stringify({
      id: 'user-1',
      username: 'operator',
      email: 'operator@pylon.test',
      displayName: 'Pylon Operator',
      avatarUrl: '',
      createdAt: '2026-07-01T00:00:00.000Z',
    }),
  )
}
