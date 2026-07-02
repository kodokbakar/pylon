import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { StatusIndicator } from '../StatusIndicator'

describe('StatusIndicator', () => {
  it.each([
    ['online', 'User is online'],
    ['offline', 'User is offline'],
    ['typing', 'User is typing'],
  ] as const)('renders accessible label for %s status', (status, label) => {
    render(<StatusIndicator status={status} />)

    expect(screen.getByLabelText(label)).toBeInTheDocument()
  })

  it('uses a custom label when provided', () => {
    render(<StatusIndicator label="3 online in Engineering" status="online" />)

    expect(screen.getByLabelText('3 online in Engineering')).toBeInTheDocument()
    expect(screen.getByTitle('3 online in Engineering')).toBeInTheDocument()
  })
})
