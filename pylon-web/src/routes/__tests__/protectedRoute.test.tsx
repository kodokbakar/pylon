import { describe, expect, it } from 'vitest'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { render, screen } from '@testing-library/react'

import { AuthProvider } from '../../context/AuthContext'
import { ProtectedRoute } from '../ProtectedRoute'
import type { StoredAuthUser } from '../../utils/authToken'

describe('ProtectedRoute', () => {
  it('redirects to login when no auth token exists', async () => {
    renderProtectedRoute('/secret')

    expect(await screen.findByText('Login page')).toBeInTheDocument()
    expect(screen.queryByText('Protected content')).not.toBeInTheDocument()
  })

  it('renders protected children when auth token exists', () => {
    window.localStorage.setItem('auth_token', 'access-token')
    window.localStorage.setItem('refresh_token', 'refresh-token')
    window.localStorage.setItem('auth_user', JSON.stringify(testUser))

    renderProtectedRoute('/secret')

    expect(screen.getByText('Protected content')).toBeInTheDocument()
    expect(screen.queryByText('Login page')).not.toBeInTheDocument()
  })
})

function renderProtectedRoute(route: string) {
  return render(
    <MemoryRouter initialEntries={[route]}>
      <AuthProvider>
        <Routes>
          <Route element={<ProtectedRoute />}>
            <Route path="/secret" element={<div>Protected content</div>} />
          </Route>

          <Route path="/login" element={<div>Login page</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
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
