import { useMemo } from 'react'

import type {
  PresenceByUserId,
  PresenceEntry,
  RoomPresenceStatus,
} from '../context/presenceContext'
import { useAuth } from './useAuth'
import { usePresence } from './usePresence'
import { useStreamPresence } from './useStreamPresence'

export type { PresenceByUserId, RoomPresenceStatus }
export type RoomPresenceEntry = PresenceEntry

export function roomPresenceQueryKey(roomId: string) {
  return ['room-presence', roomId] as const
}

export function useRoomPresence(roomId: string | undefined) {
  const { user } = useAuth()
  const currentUserId = user?.id ?? ''
  const normalizedRoomId = roomId?.trim() ?? ''
  const presence = usePresence()

  useStreamPresence(normalizedRoomId)

  const presencesByUserId = useMemo(() => {
    const nextPresencesByUserId: PresenceByUserId = {}

    for (const entry of presence.getRoomPresences(normalizedRoomId)) {
      nextPresencesByUserId[entry.userId] = entry
    }

    return nextPresencesByUserId
  }, [normalizedRoomId, presence])

  const onlineCount = useMemo(
    () => presence.getOnlineCount(normalizedRoomId),
    [normalizedRoomId, presence],
  )

  const typingUserIds = useMemo(
    () => presence.getTypingUsers(normalizedRoomId, currentUserId),
    [currentUserId, normalizedRoomId, presence],
  )

  async function sendTyping() {
    if (!normalizedRoomId) {
      return false
    }

    return presence.sendTyping(normalizedRoomId)
  }

  return {
    presencesByUserId,
    onlineCount,
    typingUserIds,
    sendTyping,
    getStatus(userId: string) {
      return presence.getStatus(normalizedRoomId, userId)
    },
  }
}

export function getPresenceStatus(
  presencesByUserId: PresenceByUserId,
  userId: string,
): RoomPresenceStatus {
  return presencesByUserId[userId]?.status ?? 'offline'
}

export function isPresenceOnline(status: RoomPresenceStatus) {
  return status === 'online' || status === 'typing'
}

export function formatPresenceStatus(status: RoomPresenceStatus) {
  switch (status) {
    case 'online':
      return 'online'
    case 'typing':
      return 'typing'
    default:
      return 'offline'
  }
}
