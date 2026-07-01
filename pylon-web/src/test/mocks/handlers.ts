import { vi } from 'vitest'

export const connectRpcHandlers = {
  auth: {
    register: vi.fn(),
    login: vi.fn(),
    refreshToken: vi.fn(),
  },
  rooms: {
    createRoom: vi.fn(),
    listRooms: vi.fn(),
    getRoom: vi.fn(),
    getRoomMembers: vi.fn(),
    leaveRoom: vi.fn(),
  },
  presence: {
    getRoomPresence: vi.fn(),
    setTyping: vi.fn(),
    streamPresence: vi.fn(),
  },
}

export function resetConnectRpcHandlers() {
  for (const group of Object.values(connectRpcHandlers)) {
    for (const handler of Object.values(group)) {
      handler.mockReset()
    }
  }
}

export function asyncIterableFrom<T>(items: T[]): AsyncIterable<T> {
  return {
    async *[Symbol.asyncIterator]() {
      for (const item of items) {
        yield item
      }
    },
  }
}
