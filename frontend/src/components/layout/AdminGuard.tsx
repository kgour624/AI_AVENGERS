import { Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'

/**
 * Guards /admin/* routes specifically.
 * WHY separate from AuthGuard: FRONTEND_SYSTEM_DESIGN.md section 15 -
 * "Admin routes - double check: route guard + API returns 403 -
 * Defense in depth." This is the route-guard half; the backend's 403
 * on every /api/v1/admin/* call is the second half, already covered
 * by api/admin.ts hitting real endpoints that the backend enforces
 * independently of whatever the frontend thinks.
 *
 * This assumes it runs INSIDE AuthGuard's <Outlet /> (see App.tsx route
 * tree), so isAuthenticated is already guaranteed true here - only
 * the admin-role check is this component's job.
 */
export function AdminGuard() {
  const isAdmin = useAuthStore((s) => s.isAdmin())

  if (!isAdmin) {
    return <Navigate to="/" replace />
  }

  return <Outlet />
}
