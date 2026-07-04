import { beforeEach, describe, expect, it, vi } from 'vitest'

import { queueTokenRefresh, resetTokenRefreshQueue } from '../tokenRefresh'

describe('token refresh queue', () => {
  beforeEach(() => {
    resetTokenRefreshQueue()
  })

  it('deduplicates concurrent refresh calls into one promise', async () => {
    const refresh = vi.fn(async () => 'next-access-token')

    const results = await Promise.all([
      queueTokenRefresh(refresh),
      queueTokenRefresh(refresh),
      queueTokenRefresh(refresh),
    ])

    expect(refresh).toHaveBeenCalledTimes(1)
    expect(results).toEqual(['next-access-token', 'next-access-token', 'next-access-token'])
  })

  it('allows a new refresh after the active refresh settles', async () => {
    const refresh = vi.fn(async () => 'next-access-token')

    await queueTokenRefresh(refresh)
    await queueTokenRefresh(refresh)

    expect(refresh).toHaveBeenCalledTimes(2)
  })

  it('shares failed refresh results with all waiting callers', async () => {
    const refresh = vi.fn(async () => null)

    const results = await Promise.all([queueTokenRefresh(refresh), queueTokenRefresh(refresh)])

    expect(refresh).toHaveBeenCalledTimes(1)
    expect(results).toEqual([null, null])
  })
})
