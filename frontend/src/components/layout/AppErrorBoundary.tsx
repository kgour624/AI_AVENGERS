import { Component, type ErrorInfo, type ReactNode } from 'react'

/**
 * Global render-error boundary.
 * Source: FRONTEND_SYSTEM_DESIGN.md section 14 -
 * "export class AppErrorBoundary extends React.Component { // Catches
 * render errors, shows fallback UI }" (was listed by name in the doc
 * but never actually implemented in any prior phase - the doc names
 * it in one line with no real class body given).
 *
 * WHY a class component, not a hook-based alternative: React does not
 * currently provide a hook equivalent to componentDidCatch/
 * getDerivedStateFromError - this is one of the few remaining cases
 * where a class component is the only option, not a stylistic choice.
 *
 * WHY this is DISTINCT from RouteError.tsx (components/layout/
 * RouteError.tsx, built in Phase 1): RouteError only catches errors
 * thrown from a route's loader or from render WITHIN that route's
 * subtree, via React Router's errorElement mechanism. It does NOT
 * catch errors thrown by code entirely outside the router's control -
 * e.g. a render error inside <QueryClientProvider> itself, or in
 * main.tsx's own render call before RouterProvider even mounts. This
 * component wraps EVERYTHING (see main.tsx), so it's the true
 * last-resort fallback; RouteError remains the more specific,
 * router-aware one for errors router error handling can already see.
 */
interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export class AppErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, error: null }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // WHY console.error explicitly rather than swallowing: matches the
    // "never silent catch" principle from utils/errors.ts's own header
    // comment (Phase 2) - an error boundary that catches and says
    // nothing is strictly worse for debugging than letting it crash,
    // since at least a crash is loud.
    console.error('AppErrorBoundary caught an error:', error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex h-screen flex-col items-center justify-center gap-4 bg-surface-base p-8 text-center">
          <p className="text-lg text-text-primary">Something went wrong.</p>
          <p className="max-w-md text-sm text-text-secondary">
            {this.state.error?.message ?? 'An unexpected error occurred.'}
          </p>
          <button
            onClick={() => window.location.reload()}
            className="rounded-md bg-brand px-4 py-2 text-sm text-white hover:bg-brand-hover"
          >
            Reload
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
