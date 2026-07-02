import type { PropsWithChildren, ReactElement } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render as testingLibraryRender, type RenderOptions } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'

import { AuthProvider } from '../context/AuthContext'

type CustomRenderOptions = Omit<RenderOptions, 'wrapper'> & {
  route?: string
  queryClient?: QueryClient
}

export function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        refetchOnWindowFocus: false,
      },
      mutations: {
        retry: false,
      },
    },
  })
}

function createWrapper(route: string, queryClient: QueryClient) {
  return function TestWrapper({ children }: PropsWithChildren) {
    return (
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <MemoryRouter initialEntries={[route]}>{children}</MemoryRouter>
        </AuthProvider>
      </QueryClientProvider>
    )
  }
}

export function renderWithProviders(ui: ReactElement, options: CustomRenderOptions = {}) {
  const { route = '/', queryClient = createTestQueryClient(), ...renderOptions } = options
  const user = userEvent.setup()

  return {
    user,
    queryClient,
    ...testingLibraryRender(ui, {
      wrapper: createWrapper(route, queryClient),
      ...renderOptions,
    }),
  }
}

export * from '@testing-library/react'
export { renderWithProviders as render }
