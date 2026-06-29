export const AUTH_TOKEN_KEY = 'auth_token'

export function getAuthToken() {
  try {
    return window.localStorage.getItem(AUTH_TOKEN_KEY)
  } catch {
    return null
  }
}

export function setAuthToken(token: string) {
  window.localStorage.setItem(AUTH_TOKEN_KEY, token)
}

export function removeAuthToken() {
  window.localStorage.removeItem(AUTH_TOKEN_KEY)
}

export function hasAuthToken() {
  return Boolean(getAuthToken())
}
