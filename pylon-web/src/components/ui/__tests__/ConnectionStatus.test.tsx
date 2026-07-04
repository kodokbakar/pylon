import { describe, expect, it } from 'vitest'

import { render, screen } from '../../../test/render'
import { ConnectionStatus } from '../ConnectionStatus'

describe('ConnectionStatus', () => {
  it('renders connected state', () => {
    render(<ConnectionStatus maxReconnectAttempts={10} reconnectAttempt={0} state="connected" />)

    expect(screen.getByLabelText('Realtime connection connected')).toHaveTextContent('Connected')
  })

  it('renders reconnecting state with attempt and next delay', () => {
    render(
      <ConnectionStatus
        maxReconnectAttempts={10}
        nextReconnectDelayMs={4_000}
        reconnectAttempt={3}
        state="reconnecting"
      />,
    )

    expect(
      screen.getByLabelText('Realtime connection reconnecting, attempt 3 of 10, next retry in 4s'),
    ).toHaveTextContent('Reconnecting 3/10 · 4s')
  })

  it('renders failed state after max attempts', () => {
    render(<ConnectionStatus maxReconnectAttempts={10} reconnectAttempt={10} state="error" />)

    expect(
      screen.getByLabelText('Realtime connection disconnected after 10 of 10 reconnect attempts'),
    ).toHaveTextContent('Disconnected 10/10')
  })
})
