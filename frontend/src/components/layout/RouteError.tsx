import { useRouteError, isRouteErrorResponse } from 'react-router-dom'

/**
 * Route-level error boundary element.
 * Source: FRONTEND_SYSTEM_DESIGN.md section 14.
 * WHY per-route rather than only a top-level React error boundary:
 * a loader throwing (e.g. getChat() 404s because the chat was deleted
 * in another tab) is a routing-layer error, not a render error - only
 * errorElement on the route catches that; a component-level
 * componentDidCatch boundary does not.
 */
export function RouteError() {
  const error = useRouteError()

  let message = 'Something went wrong.'
  if (isRouteErrorResponse(error)) {
    message = `${error.status} ${error.statusText}`
  } else if (error instanceof Error) {
    message = error.message
  }

  return (
    <div className="flex h-full flex-col items-center justify-center gap-4 p-8 text-center">
      <p className="text-lg text-text-primary">{message}</p>
      <button
        onClick={() => window.location.reload()}
        className="rounded-md bg-brand px-4 py-2 text-sm text-white hover:bg-brand-hover"
      >
        Retry
      </button>
    </div>
  )
}
