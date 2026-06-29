import { createPromiseClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'

import { getApiBaseUrl } from './config'
import { AuthService as AuthServiceDefinition } from './gen/pylon/auth/v1/auth_service_connect'
import { getRefreshToken, removeAuthSession, setAuthTokens } from '../utils/authToken'

const refreshTransport = createConnectTransport({
  baseUrl: getApiBaseUrl(),
})

const refreshClient = createPromiseClient(AuthServiceDefinition, refreshTransport)

let refreshPromise: Promise<string | null> | null = null

export function refreshAccessToken() {
  if (!refreshPromise) {
    refreshPromise = refreshAccessTokenOnce().finally(() => {
      refreshPromise = null
    })
  }

  return refreshPromise
}

export function redirectToLogin() {
  if (window.location.pathname !== '/login') {
    window.location.assign('/login')
  }
}

async function refreshAccessTokenOnce() {
  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    removeAuthSession()
    redirectToLogin()
    return null
  }

  try {
    const response = await refreshClient.refreshToken({
      refreshToken,
    })

    const nextToken = response.token.trim()
    const nextRefreshToken = response.refreshToken.trim()

    if (!nextToken || !nextRefreshToken) {
      removeAuthSession()
      redirectToLogin()
      return null
    }

    setAuthTokens(nextToken, nextRefreshToken)
    return nextToken
  } catch {
    removeAuthSession()
    redirectToLogin()
    return null
  }
}
