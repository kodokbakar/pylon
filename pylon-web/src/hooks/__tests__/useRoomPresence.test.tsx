import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { AuthProvider } from '../../context/AuthContext'
import {
  PresenceContext,
  type PresenceContextValue,
  type PresenceEntry,
} from '../../context/presenceContext'
import { useRoomPresence } from '../useRoomPresence'

describe('useRoomPresence', () => {
  it('starts stream and returns room-level derived presence state', async () => {
    const entries: PresenceEntry[] = [
      {
        userId: 'user-current',
        roomId: 'room-1',
        status: 'typing',
        lastSeen: new Date('2026-07-01T00:00:00.000Z'),
      },
      {
        userId: 'user-2',
        roomId: 'room-1',
        status: 'online',
        lastSeen: new Date('2026-07-01T00:00:00.000Z'),
      },
      {
        userId: 'user-3',
        roomId: 'room-1',
        status: 'typing',
        lastSeen: new Date('2026-07-01T00:00:00.000Z'),
      },
      {
        userId: 'user-4',
        roomId: 'room-2',
        status: 'online',
        lastSeen: new Date('2026-07-01T00:00:00.000Z'),
      },
    ]

    const presence = createPresenceContextValue({
      getRoomPresences: vi.fn((roomId: string) =>
        entries.filter((entry) => entry.roomId === roomId),
      ),
      getOnlineCount: vi.fn(
        (roomId: string) =>
          entries.filter(
            (entry) =>
              entry.roomId === roomId && (entry.status === 'online' || entry.status === 'typing'),
          ).length,
      ),
      getTypingUsers: vi.fn((roomId: string, excludeUserId = '') =>
        entries
          .filter(
            (entry) =>
              entry.roomId === roomId &&
              entry.status === 'typing' &&
              entry.userId !== excludeUserId,
          )
          .map((entry) => entry.userId),
      ),
      getStatus: vi.fn((roomId: string, userId: string) => {
        return (
          entries.find((entry) => entry.roomId === roomId && entry.userId === userId)?.status ??
          'offline'
        )
      }),
    })

    renderUseRoomPresence('room-1', presence)

    await waitFor(() => {
      expect(presence.startRoomStream).toHaveBeenCalledWith('room-1')
    })

    expect(screen.getByTestId('online-count')).toHaveTextContent('3')
    expect(screen.getByTestId('typing-users')).toHaveTextContent('user-3')
    expect(screen.getByTestId('user-2-status')).toHaveTextContent('online')
    expect(screen.getByTestId('user-4-status')).toHaveTextContent('offline')
    expect(screen.getByTestId('known-presences')).toHaveTextContent('user-current,user-2,user-3')
  })

  it('does not start stream when room id is empty', () => {
    const presence = createPresenceContextValue()

    renderUseRoomPresence('   ', presence)

    expect(presence.startRoomStream).not.toHaveBeenCalled()
    expect(screen.getByTestId('online-count')).toHaveTextContent('0')
  })

  it('calls sendTyping with the normalized room id', async () => {
    const sendTyping = vi.fn(async () => true)
    const presence = createPresenceContextValue({
      sendTyping,
    })

    renderUseRoomPresence(' room-1 ', presence)

    await screen.getByRole('button', { name: 'Send typing' }).click()

    expect(sendTyping).toHaveBeenCalledTimes(1)
    expect(sendTyping).toHaveBeenCalledWith('room-1')
    expect(await screen.findByTestId('send-typing-result')).toHaveTextContent('true')
  })
})

function renderUseRoomPresence(roomId: string | undefined, presence: PresenceContextValue) {
  storeAuthSession()

  return render(
    <AuthProvider>
      <PresenceContext.Provider value={presence}>
        <RoomPresenceProbe roomId={roomId} />
      </PresenceContext.Provider>
    </AuthProvider>,
  )
}

function RoomPresenceProbe({ roomId }: { roomId: string | undefined }) {
  const roomPresence = useRoomPresence(roomId)
  const [sendTypingResult, setSendTypingResult] = React.useState('')

  return (
    <section>
      <p data-testid="online-count">{roomPresence.onlineCount}</p>
      <p data-testid="typing-users">{roomPresence.typingUserIds.join(',')}</p>
      <p data-testid="user-2-status">{roomPresence.getStatus('user-2')}</p>
      <p data-testid="user-4-status">{roomPresence.getStatus('user-4')}</p>
      <p data-testid="known-presences">{Object.keys(roomPresence.presencesByUserId).join(',')}</p>
      <p data-testid="send-typing-result">{sendTypingResult}</p>

      <button
        type="button"
        onClick={() => {
          void roomPresence.sendTyping().then((result) => setSendTypingResult(String(result)))
        }}
      >
        Send typing
      </button>
    </section>
  )
}

function createPresenceContextValue(
  overrides: Partial<PresenceContextValue> = {},
): PresenceContextValue {
  return {
    presences: new Map(),
    startRoomStream: vi.fn<PresenceContextValue['startRoomStream']>(() => () => undefined),
    sendTyping: vi.fn<PresenceContextValue['sendTyping']>(async () => true),
    getRoomPresences: vi.fn<PresenceContextValue['getRoomPresences']>(() => []),
    getStatus: vi.fn<PresenceContextValue['getStatus']>(() => 'offline'),
    getOnlineCount: vi.fn<PresenceContextValue['getOnlineCount']>(() => 0),
    getTypingUsers: vi.fn<PresenceContextValue['getTypingUsers']>(() => []),
    isOnline: vi.fn<PresenceContextValue['isOnline']>(() => false),
    ...overrides,
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
