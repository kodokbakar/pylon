import { beforeEach, describe, expect, it, vi } from 'vitest'

import { CreateRoomResponse, ListRoomsResponse, Room } from '../gen/pylon/room/v1/room_service_pb'

const mockRoomClient = vi.hoisted(() => ({
  createRoom: vi.fn(),
  listRooms: vi.fn(),
  getRoom: vi.fn(),
  getRoomMembers: vi.fn(),
  leaveRoom: vi.fn(),
}))

vi.mock('@connectrpc/connect', () => ({
  createPromiseClient: vi.fn(() => mockRoomClient),
}))

import { RoomService } from '../rooms'

describe('RoomService API client', () => {
  beforeEach(() => {
    mockRoomClient.createRoom.mockReset()
    mockRoomClient.listRooms.mockReset()
    mockRoomClient.getRoom.mockReset()
    mockRoomClient.getRoomMembers.mockReset()
    mockRoomClient.leaveRoom.mockReset()
  })

  it('calls listRooms and returns the mocked room list response', async () => {
    const response = new ListRoomsResponse({
      rooms: [
        new Room({
          id: 'room-1',
          name: 'Engineering',
        }),
      ],
    })

    mockRoomClient.listRooms.mockResolvedValue(response)

    await expect(
      RoomService.listRooms({
        userId: 'user-1',
      }),
    ).resolves.toBe(response)

    expect(mockRoomClient.listRooms).toHaveBeenCalledTimes(1)
    expect(mockRoomClient.listRooms).toHaveBeenCalledWith({
      userId: 'user-1',
    })
    expect(response.rooms).toHaveLength(1)
    expect(response.rooms[0]?.id).toBe('room-1')
    expect(response.rooms[0]?.name).toBe('Engineering')
  })

  it('calls createRoom and returns the mocked created room response', async () => {
    const response = new CreateRoomResponse({
      room: new Room({
        id: 'room-created',
        name: 'Incident Room',
      }),
    })

    mockRoomClient.createRoom.mockResolvedValue(response)

    await expect(
      RoomService.createRoom({
        name: 'Incident Room',
      }),
    ).resolves.toBe(response)

    expect(mockRoomClient.createRoom).toHaveBeenCalledTimes(1)
    expect(mockRoomClient.createRoom).toHaveBeenCalledWith({
      name: 'Incident Room',
    })
    expect(response.room?.id).toBe('room-created')
    expect(response.room?.name).toBe('Incident Room')
  })

  it('calls leaveRoom and resolves when the mocked client resolves', async () => {
    mockRoomClient.leaveRoom.mockResolvedValue(undefined)

    await expect(
      RoomService.leaveRoom({
        roomId: 'room-1',
      }),
    ).resolves.toBeUndefined()

    expect(mockRoomClient.leaveRoom).toHaveBeenCalledTimes(1)
    expect(mockRoomClient.leaveRoom).toHaveBeenCalledWith({
      roomId: 'room-1',
    })
  })
})
