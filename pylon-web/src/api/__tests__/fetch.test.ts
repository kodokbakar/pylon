import { beforeEach, describe, expect, it, vi } from 'vitest'

const mockGetAuthToken = vi.hoisted(() => vi.fn())
const mockRemoveAuthSession = vi.hoisted(() => vi.fn())
const mockRefreshAccessToken = vi.hoisted(() => vi.fn())
const mockRedirectToLogin = vi.hoisted(() => vi.fn())

vi.mock('../../utils/authToken', () => ({
  getAuthToken: mockGetAuthToken,
  removeAuthSession: mockRemoveAuthSession,
}))

vi.mock('../authRefresh', () => ({
  refreshAccessToken: mockRefreshAccessToken,
  redirectToLogin: mockRedirectToLogin,
}))

import { apiFetch } from '../fetch'

describe('apiFetch', () => {
  beforeEach(() => {
    mockGetAuthToken.mockReset()
    mockRemoveAuthSession.mockReset()
    mockRefreshAccessToken.mockReset()
    mockRedirectToLogin.mockReset()
    vi.restoreAllMocks()
  })

  it('attaches the current access token to requests', async () => {
    mockGetAuthToken.mockReturnValue('access-token')
    const fetchMock = vi.spyOn(window, 'fetch').mockResolvedValueOnce(new Response(null))

    await apiFetch('https://api.pylon.test/v1/resource')

    const request = fetchMock.mock.calls[0]?.[0]
    expect(request).toBeInstanceOf(Request)
    expect((request as Request).headers.get('Authorization')).toBe('Bearer access-token')
  })

  it('refreshes once and retries a 401 request with the new token', async () => {
    mockGetAuthToken.mockReturnValue('old-access-token')
    mockRefreshAccessToken.mockResolvedValue('next-access-token')

    const fetchMock = vi
      .spyOn(window, 'fetch')
      .mockResolvedValueOnce(new Response(null, { status: 401 }))
      .mockResolvedValueOnce(new Response('ok', { status: 200 }))

    const response = await apiFetch('https://api.pylon.test/v1/resource')

    expect(response.status).toBe(200)
    expect(mockRefreshAccessToken).toHaveBeenCalledTimes(1)
    expect(fetchMock).toHaveBeenCalledTimes(2)

    const firstRequest = fetchMock.mock.calls[0]?.[0]
    const retryRequest = fetchMock.mock.calls[1]?.[0]

    expect(firstRequest).toBeInstanceOf(Request)
    expect(retryRequest).toBeInstanceOf(Request)
    expect((firstRequest as Request).headers.get('Authorization')).toBe('Bearer old-access-token')
    expect((retryRequest as Request).headers.get('Authorization')).toBe('Bearer next-access-token')
  })

  it('clears the session and redirects when refresh fails', async () => {
    mockGetAuthToken.mockReturnValue('old-access-token')
    mockRefreshAccessToken.mockResolvedValue(null)

    vi.spyOn(window, 'fetch').mockResolvedValueOnce(new Response(null, { status: 401 }))

    const response = await apiFetch('https://api.pylon.test/v1/resource')

    expect(response.status).toBe(401)
    expect(mockRemoveAuthSession).toHaveBeenCalledTimes(1)
    expect(mockRedirectToLogin).toHaveBeenCalledTimes(1)
  })

  it('clears the session and redirects when retry is still unauthorized', async () => {
    mockGetAuthToken.mockReturnValue('old-access-token')
    mockRefreshAccessToken.mockResolvedValue('next-access-token')

    vi.spyOn(window, 'fetch')
      .mockResolvedValueOnce(new Response(null, { status: 401 }))
      .mockResolvedValueOnce(new Response(null, { status: 401 }))

    const response = await apiFetch('https://api.pylon.test/v1/resource')

    expect(response.status).toBe(401)
    expect(mockRemoveAuthSession).toHaveBeenCalledTimes(1)
    expect(mockRedirectToLogin).toHaveBeenCalledTimes(1)
  })
})
