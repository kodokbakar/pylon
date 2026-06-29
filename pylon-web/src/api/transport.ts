import { createConnectTransport } from '@connectrpc/connect-web'

import { getApiBaseUrl } from './config'
import { apiFetch } from './fetch'

export const publicTransport = createConnectTransport({
  baseUrl: getApiBaseUrl(),
})

export const authenticatedTransport = createConnectTransport({
  baseUrl: getApiBaseUrl(),
  fetch: apiFetch,
})
