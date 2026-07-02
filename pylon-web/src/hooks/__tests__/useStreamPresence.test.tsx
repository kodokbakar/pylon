import { render } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { PresenceContext, type PresenceContextValue } from '../../context/presenceContext'
import { useStreamPresence } from '../useStreamPresence'

describe('useStreamPresence', () => {
  it('starts a room stream when mounted with a valid room id', () => {
    const cleanup = vi.fn()
    const presence = createPresenceContextValue({
      startRoomStream: vi.fn(() => cleanup),
    })

    renderUseStreamPresence(' room-1 ', presence)

    expect(presence.startRoomStream).toHaveBeenCalledTimes(1)
    expect(presence.startRoomStream).toHaveBeenCalledWith('room-1')
  })

  it('runs stream cleanup on unmount', () => {
    const cleanup = vi.fn()
    const presence = createPresenceContextValue({
      startRoomStream: vi.fn(() => cleanup),
    })

    const result = renderUseStreamPresence('room-1', presence)

    result.unmount()

    expect(cleanup).toHaveBeenCalledTimes(1)
  })

  it('does nothing when room id is empty', () => {
    const presence = createPresenceContextValue()

    renderUseStreamPresence('   ', presence)

    expect(presence.startRoomStream).not.toHaveBeenCalled()
  })
})

function renderUseStreamPresence(roomId: string | undefined, presence: PresenceContextValue) {
  return render(
    <PresenceContext.Provider value={presence}>
      <StreamPresenceProbe roomId={roomId} />
    </PresenceContext.Provider>,
  )
}

function StreamPresenceProbe({ roomId }: { roomId: string | undefined }) {
  useStreamPresence(roomId)

  return null
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
