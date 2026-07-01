import { beforeEach, describe, expect, it, vi } from 'vitest'

import {
  GetRoomPresenceResponse,
  PresenceEvent,
  PresenceStatus,
} from '../gen/pylon/presence/v1/presence_service_pb'
import { asyncIterableFrom } from '../../test/mocks/handlers'

const mockPresenceClient = vi.hoisted(() => ({
  getRoomPresence: vi.fn(),
  setTyping: vi.fn(),
  streamPresence: vi.fn(),
}))

vi.mock('@connectrpc/connect', () => ({
  createPromiseClient: vi.fn(() => mockPresenceClient),
}))

import { PresenceService } from '../presence'

describe('PresenceService API client', () => {
  beforeEach(() => {
    mockPresenceClient.getRoomPresence.mockReset()
    mockPresenceClient.setTyping.mockReset()
    mockPresenceClient.streamPresence.mockReset()
  })

  it('calls getRoomPresence and returns presences', async () => {
    const response = new GetRoomPresenceResponse({
      presences: [
        new PresenceEvent({
          userId: 'user-1',
          roomId: 'room-1',
          status: PresenceStatus.ONLINE,
        }),
      ],
    })

    mockPresenceClient.getRoomPresence.mockResolvedValue(response)

    await expect(
      PresenceService.getRoomPresence({
        roomId: 'room-1',
      }),
    ).resolves.toBe(response)

    expect(mockPresenceClient.getRoomPresence).toHaveBeenCalledTimes(1)
    expect(mockPresenceClient.getRoomPresence).toHaveBeenCalledWith({
      roomId: 'room-1',
    })
    expect(response.presences).toHaveLength(1)
    expect(response.presences[0]?.userId).toBe('user-1')
    expect(response.presences[0]?.status).toBe(PresenceStatus.ONLINE)
  })

  it('calls setTyping with room and user ids', async () => {
    mockPresenceClient.setTyping.mockResolvedValue(undefined)

    await expect(
      PresenceService.setTyping({
        roomId: 'room-1',
        userId: 'user-1',
      }),
    ).resolves.toBeUndefined()

    expect(mockPresenceClient.setTyping).toHaveBeenCalledTimes(1)
    expect(mockPresenceClient.setTyping).toHaveBeenCalledWith({
      roomId: 'room-1',
      userId: 'user-1',
    })
  })

  it('returns the streamPresence async iterable from the mocked client', async () => {
    const event = new PresenceEvent({
      userId: 'user-1',
      roomId: 'room-1',
      status: PresenceStatus.TYPING,
    })
    const stream = asyncIterableFrom([event])

    mockPresenceClient.streamPresence.mockReturnValue(stream)

    const result = PresenceService.streamPresence({
      roomId: 'room-1',
    })

    const events = []
    for await (const item of result) {
      events.push(item)
    }

    expect(mockPresenceClient.streamPresence).toHaveBeenCalledTimes(1)
    expect(mockPresenceClient.streamPresence).toHaveBeenCalledWith({ roomId: 'room-1' }, undefined)
    expect(events).toEqual([event])
  })
})
