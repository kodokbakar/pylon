import { describe, expect, it } from 'vitest'
import userEvent from '@testing-library/user-event'
import { render, screen } from '@testing-library/react'

import { AuthProvider } from '../AuthContext'
import { useAuth } from '../../hooks/useAuth'
import type { StoredAuthUser } from '../../utils/authToken'

describe('AuthProvider', () => {
  it('starts unauthenticated when no token exists', () => {
    render(
      <AuthProvider>
        <AuthProbe />
      </AuthProvider>,
    )

    expect(screen.getByTestId('auth-state')).toHaveTextContent('unauthenticated')
    expect(screen.getByTestId('auth-token')).toHaveTextContent('no-token')
    expect(screen.getByTestId('auth-user')).toHaveTextContent('no-user')
  })

  it('login stores session and toggles authenticated state', async () => {
    const user = userEvent.setup()

    render(
      <AuthProvider>
        <AuthProbe />
      </AuthProvider>,
    )

    await user.click(screen.getByRole('button', { name: 'Login' }))

    expect(screen.getByTestId('auth-state')).toHaveTextContent('authenticated')
    expect(screen.getByTestId('auth-token')).toHaveTextContent('access-token')
    expect(screen.getByTestId('auth-user')).toHaveTextContent('operator')

    expect(window.localStorage.getItem('auth_token')).toBe('access-token')
    expect(window.localStorage.getItem('refresh_token')).toBe('refresh-token')
    expect(JSON.parse(window.localStorage.getItem('auth_user') ?? '{}')).toEqual(testUser)
  })

  it('logout clears session and toggles unauthenticated state', async () => {
    const user = userEvent.setup()

    window.localStorage.setItem('auth_token', 'existing-access-token')
    window.localStorage.setItem('refresh_token', 'existing-refresh-token')
    window.localStorage.setItem('auth_user', JSON.stringify(testUser))

    render(
      <AuthProvider>
        <AuthProbe />
      </AuthProvider>,
    )

    expect(screen.getByTestId('auth-state')).toHaveTextContent('authenticated')

    await user.click(screen.getByRole('button', { name: 'Logout' }))

    expect(screen.getByTestId('auth-state')).toHaveTextContent('unauthenticated')
    expect(screen.getByTestId('auth-token')).toHaveTextContent('no-token')
    expect(screen.getByTestId('auth-user')).toHaveTextContent('no-user')

    expect(window.localStorage.getItem('auth_token')).toBeNull()
    expect(window.localStorage.getItem('refresh_token')).toBeNull()
    expect(window.localStorage.getItem('auth_user')).toBeNull()
  })
})

function AuthProbe() {
  const auth = useAuth()

  return (
    <section>
      <p data-testid="auth-state">{auth.isAuthenticated ? 'authenticated' : 'unauthenticated'}</p>
      <p data-testid="auth-token">{auth.token ?? 'no-token'}</p>
      <p data-testid="auth-user">{auth.user?.username ?? 'no-user'}</p>

      <button
        type="button"
        onClick={() =>
          auth.login({
            token: 'access-token',
            refreshToken: 'refresh-token',
            user: testUser,
          })
        }
      >
        Login
      </button>

      <button type="button" onClick={auth.logout}>
        Logout
      </button>
    </section>
  )
}

const testUser: StoredAuthUser = {
  id: 'user-1',
  username: 'operator',
  email: 'operator@pylon.test',
  displayName: 'Pylon Operator',
  avatarUrl: '',
  createdAt: '2026-07-01T00:00:00.000Z',
}
