import React from 'react'
import { act, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { AuthProvider } from '../AuthContext'
import { PresenceProvider } from '../PresenceContext'
import { PresenceEvent, PresenceStatus } from '../../api/gen/pylon/presence/v1/presence_service_pb'
import { usePresence } from '../../hooks/usePresence'

const mockPresenceService = vi.hoisted(() => ({
  getRoomPresence: vi.fn(),
  setTyping: vi.fn(),
  streamPresence: vi.fn(),
}))

vi.mock('../../api/presence', () => ({
  PresenceService: mockPresenceService,
}))

describe('PresenceProvider', () => {
  beforeEach(() => {
    mockPresenceService.getRoomPresence.mockReset()
    mockPresenceService.setTyping.mockReset()
    mockPresenceService.streamPresence.mockReset()
  })

  it('initializes with empty presence state', () => {
    renderPresenceProvider(<PresenceProbe roomId="room-1" userId="user-1" />)

    expect(screen.getByTestId('presence-size')).toHaveTextContent('0')
    expect(screen.getByTestId('status')).toHaveTextContent('offline')
    expect(screen.getByTestId('online-count')).toHaveTextContent('0')
  })

  it('loads snapshot presences when a room stream starts', async () => {
    mockPresenceService.getRoomPresence.mockResolvedValueOnce({
      presences: [
        new PresenceEvent({
          userId: 'user-1',
          roomId: 'room-1',
          status: PresenceStatus.ONLINE,
        }),
      ],
    })
    mockPresenceService.streamPresence.mockReturnValue(emptyStream())

    renderPresenceProvider(
      <>
        <StreamProbe roomId="room-1" />
        <PresenceProbe roomId="room-1" userId="user-1" />
      </>,
    )

    await waitFor(() => {
      expect(screen.getByTestId('status')).toHaveTextContent('online')
    })

    expect(screen.getByTestId('presence-size')).toHaveTextContent('1')
    expect(screen.getByTestId('online-count')).toHaveTextContent('1')
    expect(mockPresenceService.getRoomPresence).toHaveBeenCalledWith({
      roomId: 'room-1',
    })
  })

  it('updates user status from streamed presence events', async () => {
    mockPresenceService.getRoomPresence.mockResolvedValueOnce({
      presences: [
        new PresenceEvent({
          userId: 'user-1',
          roomId: 'room-1',
          status: PresenceStatus.ONLINE,
        }),
      ],
    })
    mockPresenceService.streamPresence.mockReturnValue(
      streamFrom([
        new PresenceEvent({
          userId: 'user-1',
          roomId: 'room-1',
          status: PresenceStatus.TYPING,
        }),
      ]),
    )

    renderPresenceProvider(
      <>
        <StreamProbe roomId="room-1" />
        <PresenceProbe roomId="room-1" userId="user-1" />
      </>,
    )

    await waitFor(() => {
      expect(screen.getByTestId('status')).toHaveTextContent('typing')
    })

    expect(screen.getByTestId('typing-users')).toHaveTextContent('user-1')
    expect(screen.getByTestId('online-count')).toHaveTextContent('1')
  })

  it('tracks presences independently across multiple rooms', async () => {
    mockPresenceService.getRoomPresence.mockImplementation(({ roomId }: { roomId: string }) =>
      Promise.resolve({
        presences: [
          new PresenceEvent({
            userId: roomId === 'room-a' ? 'user-a' : 'user-b',
            roomId,
            status: roomId === 'room-a' ? PresenceStatus.ONLINE : PresenceStatus.OFFLINE,
          }),
        ],
      }),
    )
    mockPresenceService.streamPresence.mockReturnValue(emptyStream())

    renderPresenceProvider(
      <>
        <StreamProbe roomId="room-a" />
        <StreamProbe roomId="room-b" />
        <PresenceProbe label="room-a" roomId="room-a" userId="user-a" />
        <PresenceProbe label="room-b" roomId="room-b" userId="user-b" />
      </>,
    )

    await waitFor(() => {
      expect(screen.getByTestId('room-a-status')).toHaveTextContent('online')
      expect(screen.getByTestId('room-b-status')).toHaveTextContent('offline')
    })

    expect(screen.getByTestId('room-a-online-count')).toHaveTextContent('1')
    expect(screen.getByTestId('room-b-online-count')).toHaveTextContent('0')
    expect(screen.getByTestId('room-a-presence-size')).toHaveTextContent('2')
  })

  it('aborts an active room stream on cleanup', async () => {
    let streamSignal: AbortSignal | undefined

    mockPresenceService.getRoomPresence.mockResolvedValueOnce({
      presences: [],
    })
    mockPresenceService.streamPresence.mockImplementation(
      (_input: unknown, options?: { signal?: AbortSignal }) => {
        streamSignal = options?.signal
        return streamUntilAbort(streamSignal)
      },
    )

    const result = renderPresenceProvider(<StreamProbe roomId="room-1" />)

    await waitFor(() => {
      expect(streamSignal).toBeDefined()
    })

    expect(streamSignal?.aborted).toBe(false)

    act(() => {
      result.unmount()
    })

    expect(streamSignal?.aborted).toBe(true)
  })
})

function renderPresenceProvider(children: React.ReactNode) {
  storeAuthSession()

  return render(
    <AuthProvider>
      <PresenceProvider>{children}</PresenceProvider>
    </AuthProvider>,
  )
}

function StreamProbe({ roomId }: { roomId: string | undefined }) {
  const { startRoomStream } = usePresence()

  React.useEffect(() => {
    if (!roomId) {
      return
    }

    return startRoomStream(roomId)
  }, [roomId, startRoomStream])

  return null
}

function PresenceProbe({
  roomId,
  userId,
  label = '',
}: {
  roomId: string
  userId: string
  label?: string
}) {
  const presence = usePresence()
  const prefix = label ? `${label}-` : ''

  return (
    <section>
      <p data-testid={`${prefix}presence-size`}>{presence.presences.size}</p>
      <p data-testid={`${prefix}status`}>{presence.getStatus(roomId, userId)}</p>
      <p data-testid={`${prefix}online-count`}>{presence.getOnlineCount(roomId)}</p>
      <p data-testid={`${prefix}typing-users`}>{presence.getTypingUsers(roomId).join(',')}</p>
    </section>
  )
}

function streamFrom(events: PresenceEvent[]): AsyncIterable<PresenceEvent> {
  return {
    async *[Symbol.asyncIterator]() {
      for (const event of events) {
        yield event
      }
    },
  }
}

function emptyStream(): AsyncIterable<PresenceEvent> {
  return streamFrom([])
}

function streamUntilAbort(signal: AbortSignal | undefined): AsyncIterable<PresenceEvent> {
  return {
    async *[Symbol.asyncIterator]() {
      await new Promise<void>((resolve) => {
        if (!signal || signal.aborted) {
          resolve()
          return
        }

        signal.addEventListener('abort', () => resolve(), { once: true })
      })

      yield* []
    },
  }
}

function storeAuthSession() {
  window.localStorage.setItem('auth_token', 'access-token')
  window.localStorage.setItem('refresh_token', 'refresh-token')
  window.localStorage.setItem(
    'auth_user',
    JSON.stringify({
      id: 'user-current',
      username: 'operator',
      email: 'operator@pylon.test',
      displayName: 'Pylon Operator',
      avatarUrl: '',
      createdAt: '2026-07-01T00:00:00.000Z',
    }),
  )
}
