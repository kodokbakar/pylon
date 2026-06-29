import { useQuery } from '@tanstack/react-query'

import { RoomService } from '../api/rooms'
import type { Room } from '../api/gen/pylon/room/v1/room_service_pb'
import { useAuth } from './useAuth'

export type RoomListItemData = {
  id: string
  name: string
  initial: string
  updatedAt: string | null
  lastMessagePreview: string | null
  unreadCount: number
}

const ROOM_STALE_TIME_MS = 30_000

export function useRooms() {
  const { user, isAuthenticated } = useAuth()
  const userId = user?.id ?? ''

  const query = useQuery({
    queryKey: ['rooms', userId],
    enabled: isAuthenticated && userId.length > 0,
    staleTime: ROOM_STALE_TIME_MS,
    queryFn: async () => {
      const response = await RoomService.listRooms({
        userId,
      })

      return response.rooms.map(toRoomListItem).sort(sortRoomsByUpdatedAtDesc)
    },
  })

  return {
    ...query,
    rooms: query.data ?? [],
    missingUser: isAuthenticated && userId.length === 0,
  }
}

function toRoomListItem(room: Room): RoomListItemData {
  return {
    id: room.id,
    name: room.name || 'Untitled room',
    initial: getRoomInitial(room.name),
    updatedAt: room.createdAt?.toDate().toISOString() ?? null,
    lastMessagePreview: null,
    unreadCount: 0,
  }
}

function sortRoomsByUpdatedAtDesc(left: RoomListItemData, right: RoomListItemData) {
  return getTime(right.updatedAt) - getTime(left.updatedAt)
}

function getTime(value: string | null) {
  if (!value) {
    return 0
  }

  return new Date(value).getTime()
}

function getRoomInitial(name: string) {
  const trimmedName = name.trim()
  if (!trimmedName) {
    return '#'
  }

  return trimmedName[0]?.toUpperCase() ?? '#'
}
