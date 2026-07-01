import { beforeEach, describe, expect, it, vi } from 'vitest'

import {
  LoginResponse,
  RefreshTokenResponse,
  RegisterResponse,
  User,
} from '../gen/pylon/auth/v1/auth_service_pb'

const mockAuthClient = vi.hoisted(() => ({
  register: vi.fn(),
  login: vi.fn(),
  refreshToken: vi.fn(),
}))

vi.mock('@connectrpc/connect', () => ({
  createPromiseClient: vi.fn(() => mockAuthClient),
}))

import { AuthService } from '../auth'

describe('AuthService API client', () => {
  beforeEach(() => {
    mockAuthClient.register.mockReset()
    mockAuthClient.login.mockReset()
    mockAuthClient.refreshToken.mockReset()
  })

  it('calls login and returns the mocked login response', async () => {
    const response = new LoginResponse({
      token: 'access-token',
      refreshToken: 'refresh-token',
      user: createAuthUser(),
    })

    mockAuthClient.login.mockResolvedValue(response)

    await expect(
      AuthService.login({
        email: 'operator@pylon.test',
        password: 'password123',
      }),
    ).resolves.toBe(response)

    expect(mockAuthClient.login).toHaveBeenCalledTimes(1)
    expect(mockAuthClient.login).toHaveBeenCalledWith({
      email: 'operator@pylon.test',
      password: 'password123',
    })
  })

  it('calls register and returns the mocked register response', async () => {
    const response = new RegisterResponse({
      token: 'registered-access-token',
      refreshToken: 'registered-refresh-token',
      user: createAuthUser({
        id: 'user-register',
        username: 'new_operator',
        email: 'new@pylon.test',
      }),
    })

    mockAuthClient.register.mockResolvedValue(response)

    await expect(
      AuthService.register({
        username: 'new_operator',
        email: 'new@pylon.test',
        password: 'password123',
      }),
    ).resolves.toBe(response)

    expect(mockAuthClient.register).toHaveBeenCalledTimes(1)
    expect(mockAuthClient.register).toHaveBeenCalledWith({
      username: 'new_operator',
      email: 'new@pylon.test',
      password: 'password123',
    })
  })

  it('calls refreshToken and returns the mocked refresh response', async () => {
    const response = new RefreshTokenResponse({
      token: 'next-access-token',
      refreshToken: 'next-refresh-token',
    })

    mockAuthClient.refreshToken.mockResolvedValue(response)

    await expect(
      AuthService.refreshToken({
        refreshToken: 'old-refresh-token',
      }),
    ).resolves.toBe(response)

    expect(mockAuthClient.refreshToken).toHaveBeenCalledTimes(1)
    expect(mockAuthClient.refreshToken).toHaveBeenCalledWith({
      refreshToken: 'old-refresh-token',
    })
  })
})

function createAuthUser(overrides: Partial<User> = {}) {
  return new User({
    id: 'user-1',
    username: 'operator',
    email: 'operator@pylon.test',
    displayName: 'Pylon Operator',
    avatarUrl: '',
    createdAt: '2026-07-01T00:00:00.000Z',
    ...overrides,
  })
}
