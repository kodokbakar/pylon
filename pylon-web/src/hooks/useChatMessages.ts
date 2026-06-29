import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'

import { ChatService, type ChatHistoryMessage } from '../api/chat'
import { useWebSocketContext } from '../contexts/webSocketContext'
import { useAuth } from './useAuth'
import type { WebSocketMessage } from '../lib/ws'

export type ChatMessageStatus = 'sent' | 'sending' | 'error'

export type ChatMessage = {
  id: string
  roomId: string
  senderId: string
  senderName: string
  senderUsername: string
  senderAvatarUrl: string
  content: string
  msgType: string
  createdAt: string
  status: ChatMessageStatus
  optimistic: boolean
}

type OlderPage = {
  messages: ChatMessage[]
  hasMore: boolean
}

const chatPageSize = 50
const optimisticDedupeWindowMs = 15_000
const emptyMessages: ChatMessage[] = []

export function chatMessagesQueryKey(roomId: string) {
  return ['chat-messages', roomId, 'latest'] as const
}

export function useChatMessages(roomId: string | undefined) {
  const normalizedRoomId = roomId?.trim() ?? ''
  const { user } = useAuth()
  const userId = user?.id ?? ''
  const { subscribe, send, isConnected, state: connectionState } = useWebSocketContext()

  const [realtimeMessages, setRealtimeMessages] = useState<ChatMessage[]>([])
  const [olderPagesByRoom, setOlderPagesByRoom] = useState<Record<string, OlderPage>>({})
  const [isLoadingOlder, setIsLoadingOlder] = useState(false)
  const [sendError, setSendError] = useState<string | null>(null)

  const historyQuery = useQuery({
    queryKey: chatMessagesQueryKey(normalizedRoomId),
    enabled: normalizedRoomId.length > 0 && userId.length > 0,
    staleTime: 15_000,
    queryFn: async () => {
      const response = await ChatService.getMessages({
        roomId: normalizedRoomId,
        limit: chatPageSize,
      })

      return {
        messages: response.messages.map(historyMessageToChatMessage).sort(sortMessagesAsc),
        hasMore: response.hasMore,
      }
    },
  })

  useEffect(() => {
    if (!normalizedRoomId || !isConnected) {
      return
    }

    send({
      type: 'join',
      room_id: normalizedRoomId,
    })

    return () => {
      send({
        type: 'leave',
        room_id: normalizedRoomId,
      })
    }
  }, [isConnected, normalizedRoomId, send])

  useEffect(() => {
    if (!normalizedRoomId) {
      return
    }

    const unsubscribeMessage = subscribe('message', (message) => {
      const chatMessage = webSocketMessageToChatMessage(message)
      if (!chatMessage || chatMessage.roomId !== normalizedRoomId) {
        return
      }

      setRealtimeMessages((currentMessages) => mergeMessages([...currentMessages, chatMessage]))
      setSendError(null)
    })

    const unsubscribeError = subscribe('error', (message) => {
      const errorMessage = webSocketErrorMessage(message)
      if (errorMessage) {
        setSendError(errorMessage)
      }
    })

    return () => {
      unsubscribeMessage()
      unsubscribeError()
    }
  }, [normalizedRoomId, subscribe])

  const historyMessages = historyQuery.data?.messages ?? emptyMessages
  const olderPage = olderPagesByRoom[normalizedRoomId]
  const olderMessages = olderPage?.messages ?? emptyMessages
  const roomRealtimeMessages = realtimeMessages.filter(
    (message) => message.roomId === normalizedRoomId,
  )

  const messages = useMemo(
    () => mergeMessages([...olderMessages, ...historyMessages, ...roomRealtimeMessages]),
    [historyMessages, olderMessages, roomRealtimeMessages],
  )

  const hasMore = olderPage?.hasMore ?? Boolean(historyQuery.data?.hasMore)

  async function loadOlder() {
    if (!normalizedRoomId || !userId || !hasMore || isLoadingOlder) {
      return
    }

    const oldestMessage = messages.find((message) => !message.optimistic)
    if (!oldestMessage) {
      return
    }

    setIsLoadingOlder(true)

    try {
      const response = await ChatService.getMessages({
        roomId: normalizedRoomId,
        limit: chatPageSize,
        beforeId: oldestMessage.id,
      })

      const olderMessages = response.messages.map(historyMessageToChatMessage)

      setOlderPagesByRoom((currentPages) => {
        const currentPage = currentPages[normalizedRoomId]

        return {
          ...currentPages,
          [normalizedRoomId]: {
            messages: mergeMessages([...(currentPage?.messages ?? []), ...olderMessages]),
            hasMore: response.hasMore,
          },
        }
      })
    } finally {
      setIsLoadingOlder(false)
    }
  }

  function sendMessage(content: string) {
    const trimmedContent = content.trim()
    if (!trimmedContent || !normalizedRoomId || !userId) {
      return false
    }

    const optimisticMessage: ChatMessage = {
      id: newOptimisticMessageId(),
      roomId: normalizedRoomId,
      senderId: userId,
      senderName: user?.username || user?.email || 'You',
      senderUsername: user?.username ?? '',
      senderAvatarUrl: '',
      content: trimmedContent,
      msgType: 'text',
      createdAt: new Date().toISOString(),
      status: 'sending',
      optimistic: true,
    }

    setRealtimeMessages((currentMessages) => mergeMessages([...currentMessages, optimisticMessage]))
    setSendError(null)

    const sent = send({
      type: 'chat.message',
      roomId: normalizedRoomId,
      content: trimmedContent,
      msgType: 'text',
    })

    if (!sent) {
      setRealtimeMessages((currentMessages) =>
        currentMessages.map((message) =>
          message.id === optimisticMessage.id ? { ...message, status: 'error' } : message,
        ),
      )
      setSendError('Realtime connection is not ready.')
      return false
    }

    return true
  }

  return {
    messages,
    currentUserId: userId,
    isLoading: historyQuery.isLoading,
    isLoadingOlder,
    hasMore,
    errorMessage: historyQuery.error instanceof Error ? historyQuery.error.message : null,
    sendError,
    connectionState,
    isConnected,
    canSend: isConnected && normalizedRoomId.length > 0 && userId.length > 0,
    loadOlder,
    sendMessage,
  }
}

function historyMessageToChatMessage(message: ChatHistoryMessage): ChatMessage {
  return {
    id: message.id,
    roomId: message.roomId,
    senderId: message.senderId,
    senderName: message.senderDisplayName || message.senderUsername || 'Unknown sender',
    senderUsername: message.senderUsername,
    senderAvatarUrl: message.senderAvatarUrl,
    content: message.content,
    msgType: message.msgType,
    createdAt: message.createdAt,
    status: 'sent',
    optimistic: false,
  }
}

function webSocketMessageToChatMessage(message: WebSocketMessage): ChatMessage | null {
  if (message.type !== 'message' && message.type !== 'chat.message') {
    return null
  }

  const roomId = stringField(message, 'room_id', 'roomId')
  const content = stringField(message, 'content')
  const messageId = stringField(message, 'message_id', 'messageId', 'id')
  const sender = recordField(message, 'sender')
  const senderId =
    stringField(sender, 'id') || stringField(message, 'sender_id', 'senderId', 'user_id', 'userId')

  if (!roomId || !content || !senderId) {
    return null
  }

  const senderUsername = stringField(sender, 'username') || stringField(message, 'sender_username')
  const senderDisplayName =
    stringField(sender, 'display_name', 'displayName') || stringField(message, 'sender_name')

  return {
    id: messageId || newOptimisticMessageId(),
    roomId,
    senderId,
    senderName: senderDisplayName || senderUsername || 'Unknown sender',
    senderUsername,
    senderAvatarUrl: stringField(sender, 'avatar_url', 'avatarUrl'),
    content,
    msgType: stringField(message, 'msg_type', 'msgType') || 'text',
    createdAt: stringField(message, 'created_at', 'createdAt') || new Date().toISOString(),
    status: 'sent',
    optimistic: false,
  }
}

function webSocketErrorMessage(message: WebSocketMessage) {
  const code = stringField(message, 'code')
  const detail = stringField(message, 'message')

  if (!code && !detail) {
    return ''
  }

  return [code, detail].filter(Boolean).join(': ')
}

function mergeMessages(messages: ChatMessage[]) {
  const byId = new Map<string, ChatMessage>()

  for (const message of messages) {
    if (!message.id) {
      continue
    }

    if (!message.optimistic) {
      removeMatchingOptimisticMessage(byId, message)
    }

    byId.set(message.id, {
      ...byId.get(message.id),
      ...message,
    })
  }

  return Array.from(byId.values()).sort(sortMessagesAsc)
}

function removeMatchingOptimisticMessage(messages: Map<string, ChatMessage>, message: ChatMessage) {
  for (const [id, candidate] of messages) {
    if (!candidate.optimistic) {
      continue
    }

    if (
      candidate.roomId === message.roomId &&
      candidate.senderId === message.senderId &&
      candidate.content === message.content &&
      Math.abs(Date.parse(candidate.createdAt) - Date.parse(message.createdAt)) <=
        optimisticDedupeWindowMs
    ) {
      messages.delete(id)
    }
  }
}

function sortMessagesAsc(left: ChatMessage, right: ChatMessage) {
  return dateValue(left.createdAt) - dateValue(right.createdAt)
}

function dateValue(value: string) {
  const time = Date.parse(value)
  return Number.isFinite(time) ? time : 0
}

function newOptimisticMessageId() {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return `optimistic-${crypto.randomUUID()}`
  }

  return `optimistic-${Date.now()}-${Math.random().toString(36).slice(2)}`
}

function recordField(record: Record<string, unknown>, ...keys: string[]) {
  const value = valueField(record, ...keys)
  return isRecord(value) ? value : {}
}

function stringField(record: Record<string, unknown>, ...keys: string[]) {
  const value = valueField(record, ...keys)
  return typeof value === 'string' ? value.trim() : ''
}

function valueField(record: Record<string, unknown>, ...keys: string[]) {
  for (const key of keys) {
    if (key in record) {
      return record[key]
    }
  }

  return undefined
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}
