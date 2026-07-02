import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { TypingIndicator } from '../TypingIndicator'

describe('TypingIndicator', () => {
  it('renders nothing when no users are typing', () => {
    const { container } = render(<TypingIndicator names={[]} />)

    expect(container).toBeEmptyDOMElement()
  })

  it('renders singular typing text', () => {
    render(<TypingIndicator names={['Alice']} />)

    expect(screen.getByText('Alice is typing')).toBeInTheDocument()
  })

  it('renders two-user typing text', () => {
    render(<TypingIndicator names={['Alice', 'Bob']} />)

    expect(screen.getByText('Alice, Bob are typing')).toBeInTheDocument()
  })

  it('summarizes more than two typing users', () => {
    render(<TypingIndicator names={['Alice', 'Bob', 'Charlie']} />)

    expect(screen.getByText('Alice, Bob, +1 are typing')).toBeInTheDocument()
  })
})
