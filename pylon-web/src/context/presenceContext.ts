import { createContext } from 'react'

export type RoomPresenceStatus = 'online' | 'offline' | 'typing'

export type PresenceEntry = {
  userId: string
  roomId: string
  status: RoomPresenceStatus
  lastSeen: Date
}

export type PresenceByUserId = Record<string, PresenceEntry>

export type PresenceStore = Map<string, PresenceEntry>

export type PresenceContextValue = {
  presences: PresenceStore
  startRoomStream: (roomId: string) => () => void
  sendTyping: (roomId: string) => Promise<boolean>
  getRoomPresences: (roomId: string) => PresenceEntry[]
  getStatus: (roomId: string, userId: string) => RoomPresenceStatus
  getOnlineCount: (roomId: string) => number
  getTypingUsers: (roomId: string, excludeUserId?: string) => string[]
  isOnline: (roomId: string, userId: string) => boolean
}

export const PresenceContext = createContext<PresenceContextValue | null>(null)
