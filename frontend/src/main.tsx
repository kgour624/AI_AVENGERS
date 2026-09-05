import React from 'react'
import ReactDOM from 'react-dom/client'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from '@/api/queryKeys'
import { AppErrorBoundary } from '@/components/layout/AppErrorBoundary'
import { App } from './App'
import './index.css'

// WHY import queryClient from api/queryKeys.ts instead of constructing
// one here: a client defined inline in this file cannot be imported by
// non-component code (e.g. a logout handler calling queryClient.clear()
// from stores/authStore.ts) without creating a SECOND QueryClient
// instance - which would silently break caching/deduplication, since
// TanStack Query's cache is scoped per-instance. Caught this by tracing
// through "what if something outside a component needs the client"
// before finalizing this file.

const rootEl = document.getElementById('root')
if (!rootEl) {
  throw new Error('Root element #root not found in index.html')
}

// WHY AppErrorBoundary wraps QueryClientProvider (outermost), not the
// other way around: per its own header comment, this is the
// true last-resort catch-all - it needs to be able to catch an error
// thrown from anywhere inside, including a hypothetical failure in
// QueryClientProvider's own render, not just from App/RouterProvider
// downward.
ReactDOM.createRoot(rootEl).render(
  <React.StrictMode>
    <AppErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    </AppErrorBoundary>
  </React.StrictMode>
)
