import React from 'react'
import ReactDOM from 'react-dom/client'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from '@/api/queryKeys'
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

ReactDOM.createRoot(rootEl).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  </React.StrictMode>
)
