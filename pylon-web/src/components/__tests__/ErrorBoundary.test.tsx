import { describe, expect, it, vi } from 'vitest'

import { render, screen } from '../../test/render'
import { ErrorBoundary } from '../ErrorBoundary'

describe('ErrorBoundary', () => {
  it('renders children when the tree is healthy', () => {
    render(
      <ErrorBoundary>
        <p>Healthy panel</p>
      </ErrorBoundary>,
    )

    expect(screen.getByText('Healthy panel')).toBeInTheDocument()
  })

  it('renders a generic fallback without leaking the raw error message', () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)

    render(
      <ErrorBoundary>
        <BrokenPanel message="WebSocket provider exploded" />
      </ErrorBoundary>,
    )

    expect(screen.getByRole('alert')).toBeInTheDocument()
    expect(
      screen.getByText('This section stopped rendering. Please retry or return home.'),
    ).toBeInTheDocument()
    expect(screen.queryByText('WebSocket provider exploded')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: /go home/i })).toBeInTheDocument()
  })

  it('resets and remounts children from the fallback button', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)

    let shouldThrow = true

    function RecoverablePanel() {
      if (shouldThrow) {
        throw new Error('Presence provider failed')
      }

      return <p>Presence provider recovered</p>
    }

    const { user } = render(
      <ErrorBoundary>
        <RecoverablePanel />
      </ErrorBoundary>,
    )

    expect(screen.getByText('This section stopped rendering. Please retry or return home.')).toBeInTheDocument()
    expect(screen.queryByText('Presence provider failed')).not.toBeInTheDocument()

    shouldThrow = false
    await user.click(screen.getByRole('button', { name: /try again/i }))

    expect(await screen.findByText('Presence provider recovered')).toBeInTheDocument()
  })

  it('calls the home navigation handler from the fallback', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)

    const onNavigateHome = vi.fn()
    const { user } = render(
      <ErrorBoundary onNavigateHome={onNavigateHome}>
        <BrokenPanel message="Route crashed" />
      </ErrorBoundary>,
    )

    await user.click(screen.getByRole('button', { name: /go home/i }))

    expect(onNavigateHome).toHaveBeenCalledWith('/')
  })
})

function BrokenPanel({ message }: { message: string }): never {
  throw new Error(message)
}
