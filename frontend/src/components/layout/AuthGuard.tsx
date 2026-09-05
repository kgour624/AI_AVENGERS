import { Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { useAuthBootstrap } from '@/hooks/useAuth'

/**
 * Protected-route wrapper. Source: FRONTEND_SYSTEM_DESIGN.md section 3
 * (components/layout/AuthGuard.tsx).
 *
 * Cross-question: what renders while we don't yet know if the user is
 * authenticated (bootstrap refresh in flight)? Rendering <Navigate
 * to="/login"> immediately would flash a login-page redirect for every
 * legitimately-logged-in user on every reload, then bounce back once
 * the bootstrap resolves - a jarring UX bug. Fixed by gating on
 * isBootstrapping first and rendering nothing (or a spinner) until
 * that resolves, only THEN deciding whether to redirect.
 */
export function AuthGuard() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated())
  const { isBootstrapping } = useAuthBootstrap()

  if (isBootstrapping) {
    return (
      <div className="flex h-screen items-center justify-center bg-surface-base text-text-secondary">
        Loading...
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <Outlet />
}
