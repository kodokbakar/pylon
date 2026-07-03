import { getAuthToken, removeAuthSession } from '../utils/authToken'
import { redirectToLogin, refreshAccessToken } from './authRefresh'

export async function apiFetch(input: RequestInfo | URL, init?: RequestInit) {
  const request = withAuthorization(new Request(input, init))
  const retrySource = request.clone()

  const response = await window.fetch(request)
  if (response.status !== 401) {
    return response
  }

  const nextToken = await refreshAccessToken()
  if (!nextToken) {
    removeAuthSession()
    redirectToLogin()
    return response
  }

  const retryResponse = await window.fetch(withAuthorization(retrySource, nextToken))

  if (retryResponse.status === 401) {
    removeAuthSession()
    redirectToLogin()
  }

  return retryResponse
}

function withAuthorization(request: Request, overrideToken?: string) {
  const token = overrideToken ?? getAuthToken()
  if (!token) {
    return request
  }

  const headers = new Headers(request.headers)
  headers.set('Authorization', `Bearer ${token}`)

  return new Request(request, {
    headers,
  })
}
