import { useQuery } from '@tanstack/react-query'

import { RoomService } from '../api/rooms'
import type { RoomMember } from '../api/gen/pylon/room/v1/room_service_pb'
import { getInitial } from '../utils/format'

export type RoomMemberListItem = {
  id: string
  name: string
  username: string
  role: string
  initial: string
  avatarUrl: string
  joinedAt: string | null
}

const ROOM_MEMBERS_STALE_TIME_MS = 30_000

export function roomMembersQueryKey(roomId: string) {
  return ['room-members', roomId] as const
}

export function useRoomMembers(roomId: string | undefined) {
  const normalizedRoomId = roomId?.trim() ?? ''

  const query = useQuery({
    queryKey: roomMembersQueryKey(normalizedRoomId),
    enabled: normalizedRoomId.length > 0,
    staleTime: ROOM_MEMBERS_STALE_TIME_MS,
    queryFn: async () => {
      const response = await RoomService.getRoomMembers({
        roomId: normalizedRoomId,
      })

      return response.members.map(toRoomMemberListItem).sort(sortMembers)
    },
  })

  return {
    ...query,
    members: query.data ?? [],
  }
}

function toRoomMemberListItem(member: RoomMember): RoomMemberListItem {
  const username = member.username.trim()
  const displayName = member.displayName.trim()
  const name = displayName || username || 'Unknown member'

  return {
    id: member.userId,
    name,
    username,
    role: member.role.trim() || 'member',
    initial: getInitial(name),
    avatarUrl: member.avatarUrl.trim(),
    joinedAt: member.joinedAt?.toDate().toISOString() ?? null,
  }
}

function sortMembers(left: RoomMemberListItem, right: RoomMemberListItem) {
  return roleWeight(left.role) - roleWeight(right.role) || left.name.localeCompare(right.name)
}

function roleWeight(role: string) {
  switch (role.toLowerCase()) {
    case 'owner':
      return 0
    case 'admin':
      return 1
    default:
      return 2
  }
}
