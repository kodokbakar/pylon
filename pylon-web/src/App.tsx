import { createBrowserRouter, RouterProvider } from 'react-router-dom'

import { WebSocketProvider } from './context/WebSocketContext'
import { HomePage } from './pages/HomePage'
import { LoginPage } from './pages/Login'
import { NotFoundPage } from './pages/NotFoundPage'
import { RegisterPage } from './pages/Register'
import { ChatPage } from './pages/ChatPage'
import { RoomDetailPage } from './pages/RoomDetailPage'
import { ProtectedRoute } from './routes/ProtectedRoute'
import { PublicRoute } from './routes/PublicRoute'
import { PresenceProvider } from './context/PresenceContext'
import type { PropsWithChildren } from 'react'
import { AppLayout } from './components/layout/AppLayout'

import { useAuth } from './hooks/useAuth'

const router = createBrowserRouter([
  {
    element: <PublicRoute />,
    children: [
      {
        path: '/login',
        element: <LoginPage />,
      },
      {
        path: '/register',
        element: <RegisterPage />,
      },
    ],
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <AppLayout />,
        children: [
          {
            path: '/',
            element: <HomePage />,
          },
          {
            path: '/rooms/:roomId',
            element: <RoomDetailPage />,
          },
          {
            path: '/rooms/:roomId/chat',
            element: <ChatPage />,
          },
        ],
      },
    ],
  },
  {
    path: '*',
    element: <NotFoundPage />,
  },
])

export default function App() {
  return (
    <WebSocketProvider>
      <PresenceProviderScope>
        <RouterProvider router={router} />
      </PresenceProviderScope>
    </WebSocketProvider>
  )
}

function PresenceProviderScope({ children }: PropsWithChildren) {
  const { token, user } = useAuth()
  const providerKey = token ? (user?.id ?? token) : 'anonymous'

  return <PresenceProvider key={providerKey}>{children}</PresenceProvider>
}
