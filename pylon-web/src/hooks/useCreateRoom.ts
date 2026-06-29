import { useMutation, useQueryClient } from '@tanstack/react-query'

import { RoomService } from '../api/rooms'
import { RoomType, type Room } from '../api/gen/pylon/room/v1/room_service_pb'
import {
  roomsQueryKey,
  sortRoomsByUpdatedAtDesc,
  toRoomListItem,
  type RoomListItemData,
} from './useRooms'
import { useAuth } from './useAuth'

export type CreateRoomInput = {
  name: string
  description?: string
}

export function useCreateRoom() {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const userId = user?.id ?? ''

  return useMutation({
    mutationFn: async (input: CreateRoomInput) => {
      if (!userId) {
        throw new Error('Cannot create room without an authenticated user.')
      }

      const response = await RoomService.createRoom({
        name: input.name.trim(),
        type: RoomType.GROUP,
        creatorId: userId,
        memberIds: [userId],
      })

      if (!response.room || response.room.id.trim() === '') {
        throw new Error('Room service did not return a created room.')
      }

      return response.room
    },
    onSuccess: (room) => {
      queryClient.setQueryData<RoomListItemData[]>(roomsQueryKey(userId), (currentRooms = []) => {
        const createdRoom = toRoomListItem(room)
        const nextRooms = currentRooms.filter((roomItem) => roomItem.id !== createdRoom.id)

        return [createdRoom, ...nextRooms].sort(sortRoomsByUpdatedAtDesc)
      })

      void queryClient.invalidateQueries({
        queryKey: roomsQueryKey(userId),
      })
    },
  })
}

export type CreatedRoom = Room
