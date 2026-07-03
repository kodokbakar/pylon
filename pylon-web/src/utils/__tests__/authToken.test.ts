import { describe, expect, it } from 'vitest'

import {
  getAuthToken,
  getAuthUser,
  getRefreshToken,
  getTokenRefreshDelayMs,
  isTokenExpired,
  removeAuthSession,
  setAuthSession,
  setAuthToken,
  setAuthTokens,
  setAuthUser,
  setRefreshToken,
  type StoredAuthUser,
} from '../authToken'

describe('authToken storage helpers', () => {
  it('stores and reads access token', () => {
    setAuthToken('access-token')

    expect(getAuthToken()).toBe('access-token')
  })

  it('stores and reads refresh token', () => {
    setRefreshToken('refresh-token')

    expect(getRefreshToken()).toBe('refresh-token')
  })

  it('stores a complete auth session', () => {
    setAuthSession({
      token: 'access-token',
      refreshToken: 'refresh-token',
      user: testUser,
    })

    expect(getAuthToken()).toBe('access-token')
    expect(getRefreshToken()).toBe('refresh-token')
    expect(getAuthUser()).toEqual(testUser)
  })

  it('updates only auth tokens without touching stored user', () => {
    setAuthSession({
      token: 'old-access-token',
      refreshToken: 'old-refresh-token',
      user: testUser,
    })

    setAuthTokens('next-access-token', 'next-refresh-token')

    expect(getAuthToken()).toBe('next-access-token')
    expect(getRefreshToken()).toBe('next-refresh-token')
    expect(getAuthUser()).toEqual(testUser)
  })

  it('removes the stored user when setAuthUser receives null', () => {
    setAuthUser(testUser)
    expect(getAuthUser()).toEqual(testUser)

    setAuthUser(null)

    expect(getAuthUser()).toBeNull()
  })

  it('removes all auth session values', () => {
    setAuthSession({
      token: 'access-token',
      refreshToken: 'refresh-token',
      user: testUser,
    })

    removeAuthSession()

    expect(getAuthToken()).toBeNull()
    expect(getRefreshToken()).toBeNull()
    expect(getAuthUser()).toBeNull()
  })

  it('ignores invalid stored user payloads', () => {
    window.localStorage.setItem('auth_user', JSON.stringify({ id: 'missing-fields' }))

    expect(getAuthUser()).toBeNull()
  })
})

describe('JWT expiry helpers', () => {
  it('detects expired tokens from exp', () => {
    const token = createJwt({
      iat: 100,
      exp: 200,
    })

    expect(isTokenExpired(token, 199_000)).toBe(false)
    expect(isTokenExpired(token, 200_000)).toBe(true)
  })

  it('treats malformed tokens as expired', () => {
    expect(isTokenExpired('not-a-jwt')).toBe(true)
    expect(isTokenExpired(null)).toBe(true)
  })

  it('returns the proactive refresh delay at 80 percent of token lifetime', () => {
    const token = createJwt({
      iat: 100,
      exp: 200,
    })

    expect(getTokenRefreshDelayMs(token, 100_000)).toBe(80_000)
    expect(getTokenRefreshDelayMs(token, 180_000)).toBe(0)
  })

  it('falls back to refreshing one minute before expiry when iat is missing', () => {
    const token = createJwt({
      exp: 200,
    })

    expect(getTokenRefreshDelayMs(token, 100_000)).toBe(40_000)
  })
})

function createJwt(payload: Record<string, unknown>) {
  return `${base64UrlEncode({ alg: 'none', typ: 'JWT' })}.${base64UrlEncode(payload)}.signature`
}

function base64UrlEncode(value: Record<string, unknown>) {
  return window
    .btoa(JSON.stringify(value))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '')
}

const testUser: StoredAuthUser = {
  id: 'user-1',
  username: 'operator',
  email: 'operator@pylon.test',
  displayName: 'Pylon Operator',
  avatarUrl: '',
  createdAt: '2026-07-01T00:00:00.000Z',
}
