import { afterEach, describe, expect, it, vi } from 'vitest'

import { getApiBaseUrl, getWebSocketUrl } from '../config'

describe('API config', () => {
  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('returns localhost default when VITE_API_URL is not configured', () => {
    vi.stubEnv('VITE_API_URL', undefined)

    expect(getApiBaseUrl()).toBe('http://localhost:8080')
  })

  it('returns VITE_API_URL without trailing slashes', () => {
    vi.stubEnv('VITE_API_URL', 'https://api.pylon.test///')

    expect(getApiBaseUrl()).toBe('https://api.pylon.test')
  })

  it('builds ws URL from http API URL', () => {
    vi.stubEnv('VITE_API_URL', 'http://localhost:8080')

    expect(getWebSocketUrl('access-token')).toBe('ws://localhost:8080/ws?token=access-token')
  })

  it('builds wss URL from https API URL', () => {
    vi.stubEnv('VITE_API_URL', 'https://api.pylon.test')

    expect(getWebSocketUrl('access token')).toBe('wss://api.pylon.test/ws?token=access+token')
  })
})
