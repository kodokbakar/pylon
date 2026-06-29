const AUTH_TOKEN_KEY = 'auth_token'
const REFRESH_TOKEN_KEY = 'refresh_token'
const AUTH_USER_KEY = 'auth_user'

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
