import { describe, expect, it } from 'vitest'

import { render, screen } from './render'

describe('test setup', () => {
  it('renders with testing-library and jest-dom matchers', () => {
    render(<div>Pylon test runner is ready</div>)

    expect(screen.getByText('Pylon test runner is ready')).toBeInTheDocument()
  })
})
