import { getApiBaseUrl } from './config'
import { apiFetch } from './fetch'
import { booleanField, dateField, isRecord, stringField, valueField } from '../utils/object'

export type ChatHistoryMessage = {
  id: string
  roomId: string
  senderId: string
  senderUsername: string
  senderDisplayName: string
  senderAvatarUrl: string
  content: string
  msgType: string
  createdAt: string
}

export type GetMessagesInput = {
  roomId: string
  limit?: number
  beforeId?: string
}

export type GetMessagesResult = {
  messages: ChatHistoryMessage[]
  hasMore: boolean
}

export async function getMessages(input: GetMessagesInput): Promise<GetMessagesResult> {
  const roomId = input.roomId.trim()
  if (!roomId) {
    throw new Error('Room id is required.')
  }

  const url = new URL(`/api/v1/rooms/${encodeURIComponent(roomId)}/messages`, getApiBaseUrl())
  url.searchParams.set('limit', String(input.limit ?? 50))

  const beforeId = input.beforeId?.trim()
  if (beforeId) {
    url.searchParams.set('before_id', beforeId)
  }

  const response = await apiFetch(url)
  const payload: unknown = await readJson(response)

  if (!response.ok) {
    throw new Error(getErrorMessage(payload) || 'Failed to load chat messages.')
  }

  const data = unwrapResponseData(payload)
  if (!isRecord(data)) {
    return {
      messages: [],
      hasMore: false,
    }
  }

  const rawMessages = Array.isArray(data.messages) ? data.messages : []
  const messages = rawMessages
    .map(normalizeHistoryMessage)
    .filter((message): message is ChatHistoryMessage => message !== null)

  return {
    messages,
    hasMore: booleanField(data, 'has_more', 'hasMore'),
  }
}

export const ChatService = {
  getMessages,
}

async function readJson(response: Response) {
  const text = await response.text()
  if (!text) {
    return null
  }

  try {
    return JSON.parse(text) as unknown
  } catch {
    return text
  }
}

function unwrapResponseData(payload: unknown) {
  if (isRecord(payload) && 'data' in payload) {
    return payload.data
  }

  return payload
}

function normalizeHistoryMessage(raw: unknown): ChatHistoryMessage | null {
  if (!isRecord(raw)) {
    return null
  }

  const id = stringField(raw, 'id')
  const roomId = stringField(raw, 'room_id', 'roomId')
  const senderId = stringField(raw, 'sender_id', 'senderId')
  const content = stringField(raw, 'content')

  if (!id || !roomId || !senderId) {
    return null
  }

  return {
    id,
    roomId,
    senderId,
    senderUsername: stringField(raw, 'sender_username', 'senderUsername'),
    senderDisplayName: stringField(raw, 'sender_display_name', 'senderDisplayName'),
    senderAvatarUrl: stringField(raw, 'sender_avatar_url', 'senderAvatarUrl'),
    content,
    msgType: normalizeMessageType(valueField(raw, 'type', 'msg_type', 'msgType')),
    createdAt: dateField(raw, 'created_at', 'createdAt'),
  }
}

function getErrorMessage(payload: unknown) {
  const data = unwrapResponseData(payload)

  if (isRecord(data)) {
    return stringField(data, 'message', 'error')
  }

  if (isRecord(payload)) {
    return stringField(payload, 'message', 'error')
  }

  return ''
}

function normalizeMessageType(value: unknown) {
  if (typeof value === 'string') {
    const normalized = value.trim().toLowerCase()
    if (normalized.startsWith('message_type_')) {
      return normalized.replace('message_type_', '')
    }

    return normalized || 'text'
  }

  if (typeof value === 'number') {
    switch (value) {
      case 2:
        return 'image'
      case 3:
        return 'system'
      case 4:
        return 'file'
      default:
        return 'text'
    }
  }

  return 'text'
}
