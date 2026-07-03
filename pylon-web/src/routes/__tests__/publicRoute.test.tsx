import { describe, expect, it } from 'vitest'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { render, screen } from '@testing-library/react'

import { AuthProvider } from '../../context/AuthContext'
import type { StoredAuthUser } from '../../utils/authToken'
import { PublicRoute } from '../PublicRoute'

describe('PublicRoute', () => {
  it('renders public children when unauthenticated', () => {
    renderPublicRoute('/login')

    expect(screen.getByText('Login page')).toBeInTheDocument()
    expect(screen.queryByText('Protected home')).not.toBeInTheDocument()
  })

  it('redirects authenticated users back to root', async () => {
    window.localStorage.setItem('auth_token', 'access-token')
    window.localStorage.setItem('refresh_token', 'refresh-token')
    window.localStorage.setItem('auth_user', JSON.stringify(testUser))

    renderPublicRoute('/login')

    expect(await screen.findByText('Protected home')).toBeInTheDocument()
    expect(screen.queryByText('Login page')).not.toBeInTheDocument()
  })
})

function renderPublicRoute(route: string) {
  return render(
    <MemoryRouter initialEntries={[route]}>
      <AuthProvider>
        <Routes>
          <Route element={<PublicRoute />}>
            <Route path="/login" element={<div>Login page</div>} />
          </Route>

          <Route path="/" element={<div>Protected home</div>} />
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
