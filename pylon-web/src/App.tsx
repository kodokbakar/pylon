import type { PropsWithChildren, ReactNode } from 'react'
import { createBrowserRouter, Outlet, RouterProvider, useLocation } from 'react-router-dom'

import { ErrorBoundary } from './components/ErrorBoundary'
import { AppLayout } from './components/layout/AppLayout'
import { PresenceProvider } from './context/PresenceContext'
import { WebSocketProvider } from './context/WebSocketContext'
import { useAuth } from './hooks/useAuth'
import { ChatPage } from './pages/ChatPage'
import { HomePage } from './pages/HomePage'
import { LandingPage } from './pages/LandingPage'
import { LoginPage } from './pages/Login'
import { NotFoundPage } from './pages/NotFoundPage'
import { RegisterPage } from './pages/Register'
import { RoomDetailPage } from './pages/RoomDetailPage'
import { ProtectedRoute } from './routes/ProtectedRoute'
import { PublicRoute } from './routes/PublicRoute'

const router = createBrowserRouter([
  {
    path: '/',
    element: withRouteBoundary(<RootRoute />, 'Route crashed'),
    children: [
      {
        element: <ProtectedRoute />,
        children: [
          {
            element: <AppLayout />,
            children: [
              {
                index: true,
                element: withRouteBoundary(<HomePage />, 'Home crashed'),
              },
              {
                path: 'rooms/:roomId',
                element: withRouteBoundary(<RoomDetailPage />, 'Room crashed'),
              },
              {
                path: 'rooms/:roomId/chat',
                element: withRouteBoundary(<ChatPage />, 'Chat crashed'),
              },
              {
                path: '*',
                element: withRouteBoundary(<NotFoundPage />, 'Route crashed'),
              },
            ],
          },
        ],
      },
    ],
  },
  {
    element: <PublicRoute />,
    children: [
      {
        path: '/login',
        element: withRouteBoundary(<LoginPage />, 'Login crashed'),
      },
      {
        path: '/register',
        element: withRouteBoundary(<RegisterPage />, 'Register crashed'),
      },
    ],
  },
  {
    path: '*',
    element: withRouteBoundary(<NotFoundPage />, 'Route crashed'),
  },
])

export default function App() {
  return (
    <ErrorBoundary eyebrow="Realtime provider fault" title="Realtime crashed">
      <WebSocketProvider>
        <ErrorBoundary eyebrow="Presence provider fault" title="Presence crashed">
          <PresenceProviderScope>
            <ErrorBoundary eyebrow="Router fault" title="Router crashed">
              <RouterProvider router={router} />
            </ErrorBoundary>
          </PresenceProviderScope>
        </ErrorBoundary>
      </WebSocketProvider>
    </ErrorBoundary>
  )
}

function RootRoute() {
  const { isAuthenticated } = useAuth()
  const location = useLocation()

  if (!isAuthenticated && location.pathname === '/') {
    return <LandingPage />
  }

  return <Outlet />
}

function PresenceProviderScope({ children }: PropsWithChildren) {
  const { token, user } = useAuth()
  const providerKey = token ? (user?.id ?? token) : 'anonymous'

  return <PresenceProvider key={providerKey}>{children}</PresenceProvider>
}

function withRouteBoundary(element: ReactNode, title: string) {
  return (
    <ErrorBoundary eyebrow="Route boundary" title={title}>
      {element}
    </ErrorBoundary>
  )
}
