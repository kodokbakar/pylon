import { Navigate, Outlet } from 'react-router-dom'

import { hasAuthToken } from '../utils/authToken'

export function PublicRoute() {
  if (hasAuthToken()) {
    return <Navigate to="/" replace />
  }

  return <Outlet />
}
