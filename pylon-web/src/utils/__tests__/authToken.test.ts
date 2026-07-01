import { describe, expect, it } from 'vitest'

import {
  getAuthToken,
  getAuthUser,
  getRefreshToken,
  removeAuthSession,
  setAuthSession,
  setAuthToken,
  setAuthUser,
  setRefreshToken,
  setAuthTokens,
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

const testUser: StoredAuthUser = {
  id: 'user-1',
  username: 'operator',
  email: 'operator@pylon.test',
  displayName: 'Pylon Operator',
  avatarUrl: '',
  createdAt: '2026-07-01T00:00:00.000Z',
}
