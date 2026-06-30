import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'

import { PresenceService } from '../api/presence'
import {
  PresenceStatus,
  type PresenceEvent,
} from '../api/gen/pylon/presence/v1/presence_service_pb'
import { useAuth } from './useAuth'

export type RoomPresenceStatus = 'online' | 'offline' | 'typing'

export type RoomPresenceEntry = {
  userId: string
  roomId: string
  status: RoomPresenceStatus
  timestamp: string | null
}

export type PresenceByUserId = Record<string, RoomPresenceEntry>

type StreamPresenceState = {
  roomId: string
  presencesByUserId: PresenceByUserId
}

const PRESENCE_STALE_TIME_MS = 10_000

export function roomPresenceQueryKey(roomId: string) {
  return ['room-presence', roomId] as const
}

export function useRoomPresence(roomId: string | undefined) {
  const { isAuthenticated } = useAuth()
  const normalizedRoomId = roomId?.trim() ?? ''
  const [streamState, setStreamState] = useState<StreamPresenceState>({
    roomId: '',
    presencesByUserId: {},
  })

  const query = useQuery({
    queryKey: roomPresenceQueryKey(normalizedRoomId),
    enabled: isAuthenticated && normalizedRoomId.length > 0,
    staleTime: PRESENCE_STALE_TIME_MS,
    queryFn: async () => {
      const response = await PresenceService.getRoomPresence({
        roomId: normalizedRoomId,
      })

      return toPresenceMap(response.presences)
    },
  })

  useEffect(() => {
    if (!isAuthenticated || normalizedRoomId.length === 0) {
      return
    }

    const abortController = new AbortController()

    async function readPresenceStream() {
      try {
        const stream = PresenceService.streamPresence(
          {
            roomId: normalizedRoomId,
          },
          {
            signal: abortController.signal,
          },
        )

        for await (const event of stream) {
          if (abortController.signal.aborted) {
            return
          }

          const entry = toPresenceEntry(event)
          if (!entry || entry.roomId !== normalizedRoomId) {
            continue
          }

          setStreamState((current) => {
            const currentPresences =
              current.roomId === normalizedRoomId ? current.presencesByUserId : {}

            return {
              roomId: normalizedRoomId,
              presencesByUserId: {
                ...currentPresences,
                [entry.userId]: entry,
              },
            }
          })
        }
      } catch {
        // Presence intentionally falls back to the last snapshot and offline defaults.
      }
    }

    void readPresenceStream()

    return () => {
      abortController.abort()
    }
  }, [isAuthenticated, normalizedRoomId])

  const presencesByUserId = useMemo(() => {
    const fetchedPresences = query.data ?? {}
    const streamedPresences =
      streamState.roomId === normalizedRoomId ? streamState.presencesByUserId : {}

    return {
      ...fetchedPresences,
      ...streamedPresences,
    }
  }, [normalizedRoomId, query.data, streamState])

  const onlineCount = useMemo(
    () =>
      Object.values(presencesByUserId).filter((presence) => isPresenceOnline(presence.status))
        .length,
    [presencesByUserId],
  )

  return {
    ...query,
    presencesByUserId,
    onlineCount,
    getStatus(userId: string) {
      return getPresenceStatus(presencesByUserId, userId)
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

function toPresenceMap(events: PresenceEvent[]) {
  const presencesByUserId: PresenceByUserId = {}

  for (const event of events) {
    const entry = toPresenceEntry(event)
    if (!entry) {
      continue
    }

    presencesByUserId[entry.userId] = entry
  }

  return presencesByUserId
}

function toPresenceEntry(event: PresenceEvent): RoomPresenceEntry | null {
  const userId = event.userId.trim()
  if (!userId) {
    return null
  }

  return {
    userId,
    roomId: event.roomId.trim(),
    status: protoStatusToPresenceStatus(event.status),
    timestamp: event.timestamp?.toDate().toISOString() ?? null,
  }
}

function protoStatusToPresenceStatus(status: PresenceStatus): RoomPresenceStatus {
  switch (status) {
    case PresenceStatus.ONLINE:
      return 'online'
    case PresenceStatus.TYPING:
      return 'typing'
    case PresenceStatus.OFFLINE:
    case PresenceStatus.UNSPECIFIED:
    default:
      return 'offline'
  }
}
