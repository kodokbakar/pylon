import { createConnectTransport } from '@connectrpc/connect-web'

import { getApiBaseUrl } from './config'
import { apiFetch } from './fetch'

export const publicTransport = createConnectTransport({
  baseUrl: getApiBaseUrl(),
})

// ponytail: kept for upcoming authenticated service clients; it wires token attachment and refresh.
export const authenticatedTransport = createConnectTransport({
  baseUrl: getApiBaseUrl(),
  fetch: apiFetch,
})
