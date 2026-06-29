import { useCallback } from 'react'

import { removeAuthToken, setAuthToken } from '../utils/authToken'

export function useAuth() {
  const storeToken = useCallback((token: string) => {
    setAuthToken(token)
  }, [])

  const clearToken = useCallback(() => {
    removeAuthToken()
  }, [])

  return {
    storeToken,
    clearToken,
  }
}
