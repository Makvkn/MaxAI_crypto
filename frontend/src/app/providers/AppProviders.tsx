import { useState, type ReactNode } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { env } from '@/app/config/env'
import { createQueryClient } from '@/lib/query/queryClient'
import { SessionProvider } from '@/features/auth/SessionProvider'

/**
 * Application providers.
 *
 * Server state (TanStack Query) is established first, then the session, which
 * itself is server state resolved through a query.
 */
export function AppProviders({ children }: { children: ReactNode }) {
  const [queryClient] = useState(createQueryClient)

  return (
    <QueryClientProvider client={queryClient}>
      <SessionProvider>{children}</SessionProvider>
      {env.isDev ? <ReactQueryDevtools buttonPosition="bottom-left" /> : null}
    </QueryClientProvider>
  )
}
