import { describe, expect, it } from 'vitest'

import { render } from '../../../test/render'
import { Skeleton } from '../Skeleton'

describe('Skeleton', () => {
  it('renders a reusable pulse placeholder with custom sizing', () => {
    const { container } = render(<Skeleton className="h-8 w-24" />)
    const skeleton = container.firstElementChild

    expect(skeleton).toHaveClass('animate-pulse')
    expect(skeleton).toHaveClass('bg-[var(--color-grid)]')
    expect(skeleton).toHaveClass('h-8')
    expect(skeleton).toHaveClass('w-24')
    expect(skeleton).toHaveAttribute('aria-hidden', 'true')
  })
})
