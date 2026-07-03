import { describe, expect, it } from 'vitest'

import { render, screen } from '../../test/render'
import { LandingPage } from '../LandingPage'

describe('LandingPage', () => {
  it('renders public Pylon branding and CTA links', () => {
    render(<LandingPage />)

    expect(screen.getByRole('heading', { name: 'Pylon' })).toBeInTheDocument()
    expect(
      screen.getByText(/Realtime rooms, presence, and WebSocket messaging/i),
    ).toBeInTheDocument()

    expect(screen.getAllByRole('link', { name: /login/i })[0]).toHaveAttribute('href', '/login')
    expect(screen.getAllByRole('link', { name: /register/i })[0]).toHaveAttribute(
      'href',
      '/register',
    )
  })

  it('renders the main feature sections', () => {
    render(<LandingPage />)

    expect(screen.getByRole('heading', { name: 'Realtime Chat' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Rooms' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Presence' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Microservices' })).toBeInTheDocument()
  })
})
