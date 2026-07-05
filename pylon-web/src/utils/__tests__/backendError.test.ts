import { Code, ConnectError } from '@connectrpc/connect'
import { describe, expect, it } from 'vitest'

import { cleanBackendMessage, sanitizeErrorMessage } from '../backendError'

describe('backendError utilities', () => {
  it('cleans backend validation prefixes', () => {
    expect(cleanBackendMessage('invalid input: room name is required')).toBe(
      'room name is required',
    )
  })

  it('does not expose raw plain Error messages', () => {
    expect(
      sanitizeErrorMessage(
        new Error('/internal/path/token.go: failed with secret abc123'),
        'Safe fallback.',
      ),
    ).toBe('Safe fallback.')
  })

  it('maps ConnectError codes to safe user-facing messages', () => {
    expect(
      sanitizeErrorMessage(
        new ConnectError('internal detail should not render', Code.PermissionDenied),
      ),
    ).toBe('You do not have permission to perform this action.')
  })
})
