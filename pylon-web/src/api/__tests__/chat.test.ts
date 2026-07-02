import { beforeEach, describe, expect, it, vi } from 'vitest'

const mockApiFetch = vi.hoisted(() => vi.fn())

vi.mock('../fetch', () => ({
  apiFetch: mockApiFetch,
}))

import { ChatService } from '../chat'

describe('ChatService REST client', () => {
  beforeEach(() => {
    mockApiFetch.mockReset()
  })

  it('calls the messages endpoint and normalizes chat history', async () => {
    mockApiFetch.mockResolvedValue(
      jsonResponse({
        data: {
          messages: [
            {
              id: 'message-1',
              room_id: 'room-1',
              sender_id: 'user-1',
              sender_username: 'operator',
              sender_display_name: 'Pylon Operator',
              sender_avatar_url: '',
              content: 'Hello Pylon',
              msg_type: 'MESSAGE_TYPE_TEXT',
              created_at: {
                seconds: 1_800_000_000,
                nanos: 250_000_000,
              },
            },
          ],
          has_more: true,
        },
      }),
    )

    const result = await ChatService.getMessages({
      roomId: 'room-1',
      limit: 25,
      beforeId: 'message-before',
    })

    expect(mockApiFetch).toHaveBeenCalledTimes(1)

    const requestUrl = mockApiFetch.mock.calls[0]?.[0]
    expect(requestUrl).toBeInstanceOf(URL)
    expect((requestUrl as URL).pathname).toBe('/api/v1/rooms/room-1/messages')
    expect((requestUrl as URL).searchParams.get('limit')).toBe('25')
    expect((requestUrl as URL).searchParams.get('before_id')).toBe('message-before')

    expect(result.hasMore).toBe(true)
    expect(result.messages).toEqual([
      {
        id: 'message-1',
        roomId: 'room-1',
        senderId: 'user-1',
        senderUsername: 'operator',
        senderDisplayName: 'Pylon Operator',
        senderAvatarUrl: '',
        content: 'Hello Pylon',
        msgType: 'text',
        createdAt: '2027-01-15T08:00:00.250Z',
      },
    ])
  })

  it.each([
    [2, 'image'],
    [3, 'system'],
    [4, 'file'],
    [99, 'text'],
  ])('normalizes numeric msg_type %i to %s', async (rawMsgType, expectedMsgType) => {
    mockApiFetch.mockResolvedValue(
      jsonResponse({
        data: {
          messages: [
            {
              id: `message-${rawMsgType}`,
              room_id: 'room-1',
              sender_id: 'user-1',
              content: 'Typed message',
              msg_type: rawMsgType,
              created_at: '2026-07-01T00:00:00.000Z',
            },
          ],
          has_more: false,
        },
      }),
    )

    const result = await ChatService.getMessages({
      roomId: 'room-1',
    })

    expect(result.messages[0]?.msgType).toBe(expectedMsgType)
  })

  it('uses default limit and omits before_id when beforeId is not provided', async () => {
    mockApiFetch.mockResolvedValue(
      jsonResponse({
        data: {
          messages: [],
          hasMore: false,
        },
      }),
    )

    const result = await ChatService.getMessages({
      roomId: 'room-1',
    })

    const requestUrl = mockApiFetch.mock.calls[0]?.[0]
    expect(requestUrl).toBeInstanceOf(URL)
    expect((requestUrl as URL).searchParams.get('limit')).toBe('50')
    expect((requestUrl as URL).searchParams.has('before_id')).toBe(false)

    expect(result).toEqual({
      messages: [],
      hasMore: false,
    })
  })

  it('throws when room id is empty', async () => {
    await expect(
      ChatService.getMessages({
        roomId: '   ',
      }),
    ).rejects.toThrow('Room id is required.')

    expect(mockApiFetch).not.toHaveBeenCalled()
  })

  it('throws backend error messages from failed responses', async () => {
    mockApiFetch.mockResolvedValue(
      jsonResponse(
        {
          data: {
            message: 'Room access denied.',
          },
        },
        {
          status: 403,
        },
      ),
    )

    await expect(
      ChatService.getMessages({
        roomId: 'room-1',
      }),
    ).rejects.toThrow('Room access denied.')
  })
})

function jsonResponse(payload: unknown, init: ResponseInit = {}) {
  return new Response(JSON.stringify(payload), {
    status: init.status ?? 200,
    headers: {
      'content-type': 'application/json',
      ...init.headers,
    },
  })
}
