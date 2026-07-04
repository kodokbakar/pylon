const AUTH_TOKEN_KEY = 'auth_token'
const REFRESH_TOKEN_KEY = 'refresh_token'
const AUTH_USER_KEY = 'auth_user'

const jwtRefreshRatio = 0.8
const fallbackRefreshBeforeExpiryMs = 60_000

export type StoredAuthUser = {
  id: string
  username: string
  email: string
  displayName: string
  avatarUrl: string
  createdAt: string
}

export type AuthSession = {
  token: string
  refreshToken: string
  user: StoredAuthUser | null
}

type JwtPayload = {
  exp?: number
  iat?: number
}

export function getAuthToken() {
  try {
    return window.localStorage.getItem(AUTH_TOKEN_KEY)
  } catch {
    return null
  }
}

export function getRefreshToken() {
  try {
    return window.localStorage.getItem(REFRESH_TOKEN_KEY)
  } catch {
    return null
  }
}

export function getAuthUser(): StoredAuthUser | null {
  try {
    const rawUser = window.localStorage.getItem(AUTH_USER_KEY)
    if (!rawUser) {
      return null
    }

    const parsed: unknown = JSON.parse(rawUser)
    if (!isStoredAuthUser(parsed)) {
      return null
    }

    return parsed
  } catch {
    return null
  }
}

export function setAuthToken(token: string) {
  window.localStorage.setItem(AUTH_TOKEN_KEY, token)
}

export function setRefreshToken(refreshToken: string) {
  window.localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
}

export function setAuthUser(user: StoredAuthUser | null) {
  if (!user) {
    window.localStorage.removeItem(AUTH_USER_KEY)
    return
  }

  window.localStorage.setItem(AUTH_USER_KEY, JSON.stringify(user))
}

export function setAuthSession(session: AuthSession) {
  setAuthToken(session.token)
  setRefreshToken(session.refreshToken)
  setAuthUser(session.user)
}

export function setAuthTokens(token: string, refreshToken: string) {
  setAuthToken(token)
  setRefreshToken(refreshToken)
}

export function removeAuthSession() {
  window.localStorage.removeItem(AUTH_TOKEN_KEY)
  window.localStorage.removeItem(REFRESH_TOKEN_KEY)
  window.localStorage.removeItem(AUTH_USER_KEY)
}

export function isTokenExpired(token: string | null, nowMs = Date.now()) {
  const expiresAtMs = getTokenExpiryMs(token)

  if (expiresAtMs === null) {
    return true
  }

  return expiresAtMs <= nowMs
}

export function getTokenRefreshDelayMs(token: string | null, nowMs = Date.now()) {
  const payload = decodeJwtPayload(token)
  if (!payload || typeof payload.exp !== 'number') {
    return null
  }

  const expiresAtMs = payload.exp * 1_000
  if (expiresAtMs <= nowMs) {
    return 0
  }

  if (typeof payload.iat !== 'number') {
    return Math.max(expiresAtMs - nowMs - fallbackRefreshBeforeExpiryMs, 0)
  }

  const issuedAtMs = payload.iat * 1_000
  const lifetimeMs = expiresAtMs - issuedAtMs

  if (lifetimeMs <= 0) {
    return 0
  }

  const refreshAtMs = issuedAtMs + lifetimeMs * jwtRefreshRatio
  return Math.max(Math.floor(refreshAtMs - nowMs), 0)
}

function getTokenExpiryMs(token: string | null) {
  const payload = decodeJwtPayload(token)

  if (!payload || typeof payload.exp !== 'number') {
    return null
  }

  return payload.exp * 1_000
}

function decodeJwtPayload(token: string | null): JwtPayload | null {
  if (!token) {
    return null
  }

  const [, rawPayload] = token.split('.')
  if (!rawPayload) {
    return null
  }

  try {
    const normalizedPayload = rawPayload.replace(/-/g, '+').replace(/_/g, '/')
    const paddingLength = (4 - (normalizedPayload.length % 4)) % 4
    const paddedPayload = normalizedPayload.padEnd(normalizedPayload.length + paddingLength, '=')
    const parsed: unknown = JSON.parse(window.atob(paddedPayload))

    if (!isJwtPayload(parsed)) {
      return null
    }

    return parsed
  } catch {
    return null
  }
}

function isJwtPayload(value: unknown): value is JwtPayload {
  if (typeof value !== 'object' || value === null) {
    return false
  }

  const payload = value as Record<string, unknown>

  return (
    (typeof payload.exp === 'number' || typeof payload.exp === 'undefined') &&
    (typeof payload.iat === 'number' || typeof payload.iat === 'undefined')
  )
}

function isStoredAuthUser(value: unknown): value is StoredAuthUser {
  if (typeof value !== 'object' || value === null) {
    return false
  }

  const user = value as Record<string, unknown>

  return (
    typeof user.id === 'string' &&
    typeof user.username === 'string' &&
    typeof user.email === 'string' &&
    typeof user.displayName === 'string' &&
    typeof user.avatarUrl === 'string' &&
    typeof user.createdAt === 'string'
  )
}
