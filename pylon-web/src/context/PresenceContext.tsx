import type { PropsWithChildren } from 'react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { PresenceService } from '../api/presence'
import {
  PresenceStatus,
  type PresenceEvent,
} from '../api/gen/pylon/presence/v1/presence_service_pb'
import { useAuth } from '../hooks/useAuth'
import {
  PresenceContext,
  type PresenceContextValue,
  type PresenceEntry,
  type PresenceStore,
  type RoomPresenceStatus,
} from './presenceContext'

const TYPING_EXPIRY_MS = 3_000
const TYPING_SWEEP_MS = 500

export function PresenceProvider({ children }: PropsWithChildren) {
  const { isAuthenticated } = useAuth()
  const [presences, setPresences] = useState<PresenceStore>(() => new Map())
  const roomStreamRefs = useRef(new Map<string, number>())
  const streamControllers = useRef(new Map<string, AbortController>())

  const replaceRoomPresences = useCallback((roomId: string, events: PresenceEvent[]) => {
    const normalizedRoomId = roomId.trim()
    if (!normalizedRoomId) {
      return
    }

    setPresences((currentPresences) => {
      const nextPresences = new Map(currentPresences)

      for (const [key, presence] of nextPresences) {
        if (presence.roomId === normalizedRoomId) {
          nextPresences.delete(key)
        }
      }

      for (const event of events) {
        const entry = eventToPresenceEntry(event, normalizedRoomId)
        if (!entry) {
          continue
        }

        nextPresences.set(presenceKey(entry.roomId, entry.userId), entry)
      }

      return nextPresences
    })
  }, [])

  const upsertPresence = useCallback((event: PresenceEvent) => {
    const entry = eventToPresenceEntry(event)
    if (!entry) {
      return
    }

    setPresences((currentPresences) => {
      const nextPresences = new Map(currentPresences)
      nextPresences.set(presenceKey(entry.roomId, entry.userId), entry)
      return nextPresences
    })
  }, [])

  const openRoomStream = useCallback(
    async (roomId: string) => {
      if (streamControllers.current.has(roomId)) {
        return
      }

      const abortController = new AbortController()
      streamControllers.current.set(roomId, abortController)

      try {
        const response = await PresenceService.getRoomPresence({ roomId })

        if (!abortController.signal.aborted) {
          replaceRoomPresences(roomId, response.presences)
        }

        const stream = PresenceService.streamPresence(
          { roomId },
          {
            signal: abortController.signal,
          },
        )

        for await (const event of stream) {
          if (abortController.signal.aborted) {
            return
          }

          upsertPresence(event)
        }
      } catch {
        // Presence is best-effort. Consumers default to offline when data is unavailable.
      } finally {
        if (streamControllers.current.get(roomId) === abortController) {
          streamControllers.current.delete(roomId)
        }
      }
    },
    [replaceRoomPresences, upsertPresence],
  )

  const startRoomStream = useCallback(
    (roomId: string) => {
      const normalizedRoomId = roomId.trim()
      if (!isAuthenticated || !normalizedRoomId) {
        return () => undefined
      }

      const currentRefs = roomStreamRefs.current.get(normalizedRoomId) ?? 0
      roomStreamRefs.current.set(normalizedRoomId, currentRefs + 1)

      if (currentRefs === 0) {
        void openRoomStream(normalizedRoomId)
      }

      return () => {
        const nextRefs = (roomStreamRefs.current.get(normalizedRoomId) ?? 1) - 1

        if (nextRefs > 0) {
          roomStreamRefs.current.set(normalizedRoomId, nextRefs)
          return
        }

        roomStreamRefs.current.delete(normalizedRoomId)

        const controller = streamControllers.current.get(normalizedRoomId)
        controller?.abort()
        streamControllers.current.delete(normalizedRoomId)
      }
    },
    [isAuthenticated, openRoomStream],
  )

  const sendTyping = useCallback(
    async (roomId: string) => {
      const normalizedRoomId = roomId.trim()
      if (!isAuthenticated || !normalizedRoomId) {
        return false
      }

      try {
        await PresenceService.setTyping({
          roomId: normalizedRoomId,
        })

        return true
      } catch {
        return false
      }
    },
    [isAuthenticated],
  )

  const expireTypingPresences = useCallback(() => {
    const nowMs = Date.now()

    setPresences((currentPresences) => {
      let changed = false
      const nextPresences = new Map(currentPresences)

      for (const [key, presence] of nextPresences) {
        if (!isTypingExpired(presence, nowMs)) {
          continue
        }

        nextPresences.set(key, {
          ...presence,
          status: 'online',
        })
        changed = true
      }

      return changed ? nextPresences : currentPresences
    })
  }, [])

  const hasTypingPresences = useMemo(
    () => Array.from(presences.values()).some((presence) => presence.status === 'typing'),
    [presences],
  )

  useEffect(() => {
    if (!hasTypingPresences) {
      return
    }

    const timer = window.setInterval(expireTypingPresences, TYPING_SWEEP_MS)

    return () => {
      window.clearInterval(timer)
    }
  }, [expireTypingPresences, hasTypingPresences])

  useEffect(() => {
    const activeStreamControllers = streamControllers.current
    const activeRoomStreamRefs = roomStreamRefs.current

    return () => {
      for (const controller of activeStreamControllers.values()) {
        controller.abort()
      }

      activeRoomStreamRefs.clear()
      activeStreamControllers.clear()
    }
  }, [])

  const getRoomPresences = useCallback(
    (roomId: string) => {
      const normalizedRoomId = roomId.trim()
      if (!normalizedRoomId) {
        return []
      }

      return Array.from(presences.values()).filter(
        (presence) => presence.roomId === normalizedRoomId,
      )
    },
    [presences],
  )

  const getStatus = useCallback(
    (roomId: string, userId: string): RoomPresenceStatus => {
      const presence = presences.get(presenceKey(roomId, userId))
      return presence?.status ?? 'offline'
    },
    [presences],
  )

  const isOnline = useCallback(
    (roomId: string, userId: string) => isPresenceOnline(getStatus(roomId, userId)),
    [getStatus],
  )

  const getOnlineCount = useCallback(
    (roomId: string) =>
      getRoomPresences(roomId).filter((presence) => isPresenceOnline(presence.status)).length,
    [getRoomPresences],
  )

  const getTypingUsers = useCallback(
    (roomId: string, excludeUserId = '') => {
      const normalizedExcludeUserId = excludeUserId.trim()

      return getRoomPresences(roomId)
        .filter(
          (presence) => presence.status === 'typing' && presence.userId !== normalizedExcludeUserId,
        )
        .map((presence) => presence.userId)
    },
    [getRoomPresences],
  )

  const value = useMemo<PresenceContextValue>(
    () => ({
      presences,
      startRoomStream,
      sendTyping,
      getRoomPresences,
      getStatus,
      getOnlineCount,
      getTypingUsers,
      isOnline,
    }),
    [
      getOnlineCount,
      getRoomPresences,
      getStatus,
      getTypingUsers,
      isOnline,
      presences,
      sendTyping,
      startRoomStream,
    ],
  )

  return <PresenceContext.Provider value={value}>{children}</PresenceContext.Provider>
}

function eventToPresenceEntry(event: PresenceEvent, fallbackRoomId = ''): PresenceEntry | null {
  const userId = event.userId.trim()
  const roomId = event.roomId.trim() || fallbackRoomId.trim()

  if (!userId || !roomId) {
    return null
  }

  return {
    userId,
    roomId,
    status: protoStatusToPresenceStatus(event.status),
    lastSeen: event.timestamp?.toDate() ?? new Date(),
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

function presenceKey(roomId: string, userId: string) {
  return `${roomId.trim()}:${userId.trim()}`
}

function isPresenceOnline(status: RoomPresenceStatus) {
  return status === 'online' || status === 'typing'
}

function isTypingExpired(presence: PresenceEntry, nowMs: number) {
  if (presence.status !== 'typing') {
    return false
  }

  return nowMs - presence.lastSeen.getTime() > TYPING_EXPIRY_MS
}
