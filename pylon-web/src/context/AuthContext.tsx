import type { PropsWithChildren } from 'react'
import { useCallback, useMemo, useState } from 'react'

import { AuthContext, type AuthContextValue } from './authContext'
import {
  getAuthToken,
  getAuthUser,
  removeAuthSession,
  setAuthSession,
  type AuthSession,
  type StoredAuthUser,
} from '../utils/authToken'

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
