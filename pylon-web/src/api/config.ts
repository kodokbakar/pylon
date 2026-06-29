const DEFAULT_API_URL = 'http://localhost:8080'

export function getApiBaseUrl() {
  return (import.meta.env.VITE_API_URL ?? DEFAULT_API_URL).replace(/\/+$/, '')
}

export function getWebSocketUrl(token: string) {
  const url = new URL('/ws', getApiBaseUrl())
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.searchParams.set('token', token)

  return url.toString()
}
