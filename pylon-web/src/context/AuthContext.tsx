import type { PropsWithChildren } from 'react'
import { useCallback, useEffect, useMemo, useState } from 'react'

import { refreshAccessToken } from '../api/authRefresh'
import {
  getAuthToken,
  getAuthUser,
  getTokenRefreshDelayMs,
  removeAuthSession,
  setAuthSession,
  type AuthSession,
  type StoredAuthUser,
} from '../utils/authToken'
import { AuthContext, type AuthContextValue } from './authContext'

const minProactiveRefreshDelayMs = 1_000

export function AuthProvider({ children }: PropsWithChildren) {
  const [token, setToken] = useState(() => getAuthToken())
  const [user, setUser] = useState<StoredAuthUser | null>(() => getAuthUser())

  const login = useCallback((session: AuthSession) => {
    setAuthSession(session)
    setToken(session.token)
    setUser(session.user)
  }, [])

  const logout = useCallback(() => {
    removeAuthSession()
    setToken(null)
    setUser(null)
  }, [])

  const refreshSession = useCallback(async () => {
    const nextToken = await refreshAccessToken()

    if (!nextToken) {
      setToken(null)
      setUser(null)
      return
    }

    setToken(nextToken)
  }, [])

  useEffect(() => {
    if (!token) {
      return
    }

    const refreshDelayMs = getTokenRefreshDelayMs(token)
    if (refreshDelayMs === null) {
      return
    }

    const timer = window.setTimeout(
      () => {
        void refreshSession()
      },
      Math.max(refreshDelayMs, minProactiveRefreshDelayMs),
    )

    return () => {
      window.clearTimeout(timer)
    }
  }, [refreshSession, token])

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      token,
      login,
      logout,
      isAuthenticated: Boolean(token),
    }),
    [login, logout, token, user],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
