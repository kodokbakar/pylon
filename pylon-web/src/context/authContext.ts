import { createContext } from 'react'

import type { AuthSession, StoredAuthUser } from '../utils/authToken'

export type AuthContextValue = {
  user: StoredAuthUser | null
  token: string | null
  login: (session: AuthSession) => void
  logout: () => void
  isAuthenticated: boolean
}

export const AuthContext = createContext<AuthContextValue | null>(null)
