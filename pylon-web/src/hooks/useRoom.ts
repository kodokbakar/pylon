import { Code, ConnectError } from '@connectrpc/connect'
import { useQuery } from '@tanstack/react-query'

import { RoomService } from '../api/rooms'

const ROOM_DETAIL_STALE_TIME_MS = 30_000

export function roomQueryKey(roomId: string) {
  return ['room', roomId] as const
}

export function useRoom(roomId: string | undefined) {
  const normalizedRoomId = roomId?.trim() ?? ''

  const query = useQuery({
    queryKey: roomQueryKey(normalizedRoomId),
    enabled: normalizedRoomId.length > 0,
    staleTime: ROOM_DETAIL_STALE_TIME_MS,
    queryFn: async () => {
      const response = await RoomService.getRoom({
        roomId: normalizedRoomId,
      })

      if (!response.room || response.room.id.trim() === '') {
        throw new ConnectError('room not found', Code.NotFound)
      }

      return response.room
    },
  })

  return {
    ...query,
    room: query.data ?? null,
  }
}
