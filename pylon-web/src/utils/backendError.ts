import { Code, ConnectError } from '@connectrpc/connect'

export function cleanBackendMessage(message: string) {
  return message.replace(/^invalid input:\s*/i, '').trim()
}

export function sanitizeErrorMessage(
  error: unknown,
  fallback = 'Something went wrong. Please try again.',
) {
  if (!(error instanceof ConnectError)) {
    return fallback
  }

  switch (error.code) {
    case Code.Unauthenticated:
      return 'Your session expired. Please log in again.'
    case Code.PermissionDenied:
      return 'You do not have permission to perform this action.'
    case Code.NotFound:
      return 'The requested resource could not be found.'
    case Code.InvalidArgument:
      return 'The submitted information is invalid.'
    case Code.AlreadyExists:
      return 'This item already exists.'
    case Code.Unavailable:
      return 'Service is unavailable. Please try again.'
    case Code.DeadlineExceeded:
      return 'The request timed out. Please try again.'
    default:
      return fallback
  }
}
