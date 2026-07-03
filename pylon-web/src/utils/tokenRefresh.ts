export type TokenRefreshCallback = () => Promise<string | null>

let refreshPromise: Promise<string | null> | null = null

export function queueTokenRefresh(refresh: TokenRefreshCallback) {
  if (!refreshPromise) {
    refreshPromise = refresh().finally(() => {
      refreshPromise = null
    })
  }

  return refreshPromise
}

export function resetTokenRefreshQueue() {
  refreshPromise = null
}
