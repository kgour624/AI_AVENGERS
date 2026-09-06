import { Link, useNavigate } from 'react-router-dom'
import { useUIStore } from '@/stores/uiStore'
import { useAuthStore } from '@/stores/authStore'
import { logout as logoutApi } from '@/api/auth'
import { queryClient } from '@/api/queryKeys'

/**
 * Bug 1.3 fixes (docs bug list), all in this one file:
 * 1. \u2630 and \u26A1 were bare JSX text (no quotes) - Go/JSX text
 *    nodes do NOT interpret \uXXXX escapes (that's a JS string-literal
 *    feature only), so the browser rendered the literal 6-character
 *    sequences instead of the glyphs. Wrapped in {'...'} expression
 *    containers so the JS engine actually processes the escape.
 * 2. No user name/email anywhere - added from authStore.user.
 * 3. No logout button anywhere - added, see handleLogout below.
 * 4. No link to /experts anywhere (Header or Sidebar) - added here.
 */
export function Header() {
  const navigate = useNavigate()
  const toggleSidebar = useUIStore((s) => s.toggleSidebar)
  const user = useAuthStore((s) => s.user)
  const isAdmin = useAuthStore((s) => s.isAdmin())
  const clearAuth = useAuthStore((s) => s.clearAuth)

  const handleLogout = async () => {
    try {
      // Best-effort server-side revoke (invalidates the refresh token).
      await logoutApi()
    } catch {
      // Server unreachable/already invalid - still clear locally below.
      // A failed network call must never leave the user stuck
      // "logged in" client-side with no way out.
    }
    clearAuth()
    // Purge cached query data so a next login (possibly a different
    // user, same browser) never sees a stale previous session's data.
    queryClient.clear()
    navigate('/login', { replace: true })
  }

  return (
    <header className="flex h-12 flex-shrink-0 items-center justify-between border-b border-surface-border bg-surface-raised px-4">
      <div className="flex items-center gap-4">
        <button
          onClick={toggleSidebar}
          aria-label="Toggle sidebar"
          className="text-text-secondary hover:text-text-primary"
        >
          {'\u2630'}
        </button>
        <Link to="/" className="font-semibold text-text-primary">
          {'\u26A1'} AI Avengers
        </Link>
        <Link to="/experts" className="text-sm text-text-secondary hover:text-text-primary">
          Experts
        </Link>
      </div>
      <div className="flex items-center gap-4">
        {isAdmin && (
          <Link to="/admin" className="text-sm text-text-secondary hover:text-text-primary">
            Admin
          </Link>
        )}
        {user && (
          <span className="text-sm text-text-secondary" title={user.email}>
            {user.fullName}
          </span>
        )}
        <button
          onClick={handleLogout}
          className="text-sm text-text-secondary hover:text-text-primary"
        >
          Logout
        </button>
      </div>
    </header>
  )
}
