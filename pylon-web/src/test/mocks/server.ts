import { afterEach } from 'vitest'

import { resetConnectRpcHandlers } from './handlers'

export const mockServer = {
  resetHandlers: resetConnectRpcHandlers,
  close: resetConnectRpcHandlers,
}

afterEach(() => {
  mockServer.resetHandlers()
})
