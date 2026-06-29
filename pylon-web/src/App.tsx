import { createBrowserRouter, RouterProvider } from 'react-router-dom'

import { WebSocketProvider } from './contexts/WebSocketContext'
import { HomePage } from './pages/HomePage'
import { LoginPage } from './pages/Login'
import { NotFoundPage } from './pages/NotFoundPage'
import { RegisterPage } from './pages/Register'
import { RoomDetailPage } from './pages/RoomDetailPage'
import { RoomPage } from './pages/RoomPage'
import { ProtectedRoute } from './routes/ProtectedRoute'
import { PublicRoute } from './routes/PublicRoute'

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
        path: '/',
        element: <HomePage />,
      },
      {
        path: '/rooms/:roomId',
        element: <RoomDetailPage />,
      },
      {
        path: '/rooms/:roomId/chat',
        element: <RoomPage />,
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
      <RouterProvider router={router} />
    </WebSocketProvider>
  )
}
