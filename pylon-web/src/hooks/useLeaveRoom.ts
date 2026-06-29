import { useMutation, useQueryClient } from '@tanstack/react-query'

import { RoomService } from '../api/rooms'
import { roomMembersQueryKey } from './useRoomMembers'
import { roomQueryKey } from './useRoom'
import { roomsQueryKey, type RoomListItemData } from './useRooms'
import { useAuth } from './useAuth'

export function useLeaveRoom(roomId: string | undefined) {
  const { user } = useAuth()
  const queryClient = useQueryClient()

  const normalizedRoomId = roomId?.trim() ?? ''
  const userId = user?.id ?? ''

  return useMutation({
    mutationFn: async () => {
      if (!normalizedRoomId) {
        throw new Error('Room id is required.')
      }

      if (!userId) {
        throw new Error('Cannot leave room without an authenticated user.')
      }

      await RoomService.leaveRoom({
        roomId: normalizedRoomId,
        userId,
      })
    },
    onSuccess: () => {
      queryClient.setQueryData<RoomListItemData[]>(roomsQueryKey(userId), (currentRooms = []) =>
        currentRooms.filter((room) => room.id !== normalizedRoomId),
      )

      queryClient.removeQueries({
        queryKey: roomQueryKey(normalizedRoomId),
      })

      queryClient.removeQueries({
        queryKey: roomMembersQueryKey(normalizedRoomId),
      })

      void queryClient.invalidateQueries({
        queryKey: roomsQueryKey(userId),
      })
    },
  })
}
