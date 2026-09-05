import { QueryClient } from '@tanstack/react-query'

/**
 * Shared QueryClient instance, imported by main.tsx.
 * Kept in its own file (rather than defined inline in main.tsx) so
 * non-component code - e.g. a future logout handler that needs to
 * call queryClient.clear() - can import it without importing React.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60 * 1000,
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

/**
 * Structured query keys.
 * WHY: FRONTEND_SYSTEM_DESIGN.md section 6 - "Cache invalidation is
 * predictable" when keys are built from a shared factory instead of
 * inline arrays scattered across components (a typo in an inline key
 * silently creates a new, never-invalidated cache entry).
 */
export const queryKeys = {
  experts: {
    all: ['experts'] as const,
    detail: (id: string) => ['experts', id] as const,
    topics: (id: string) => ['experts', id, 'topics'] as const,
  },
  projects: {
    all: ['projects'] as const,
    detail: (id: string) => ['projects', id] as const,
    chats: (id: string) => ['projects', id, 'chats'] as const,
    memory: (id: string) => ['projects', id, 'memory'] as const,
    timeline: (id: string) => ['projects', id, 'timeline'] as const,
  },
  chats: {
    messages: (id: string) => ['chats', id, 'messages'] as const,
  },
  admin: {
    experts: ['admin', 'experts'] as const,
    clients: ['admin', 'clients'] as const,
    stats: ['admin', 'stats'] as const,
  },
}
