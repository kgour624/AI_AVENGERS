import { useEffect, useState } from 'react'
import { useAuthStore } from '@/stores/authStore'
import { refreshSession } from '@/api/base'
import { getMe } from '@/api/auth'

/**
 * App-mount bootstrap: attempt a silent session restore via the
 * httpOnly refresh cookie before deciding whether AuthGuard should
 * redirect to /login.
 *
 * WHY this is needed at all: authStore deliberately never persists
 * accessToken (XSS mitigation - see authStore.ts). That means on
 * every hard page reload, accessToken starts out null even for a
 * legitimately logged-in user whose refresh cookie is still valid.
 * Without this hook, every reload would bounce the user to /login
 * even though their session is fine - a serious UX regression the
 * design doc's AuthGuard sketch (section 3) doesn't address at all.
 *
 * Cross-question: what if the cookie is missing/expired (user never
 * logged in, or truly logged out)? refreshSession() rejects; we catch
 * and swallow that specific case silently (it is the EXPECTED outcome
 * for a logged-out visitor, not an error to surface), and
 * isBootstrapping still flips to false so AuthGuard can redirect.
 */
export function useAuthBootstrap() {
  const accessToken = useAuthStore((s) => s.accessToken)
  const [isBootstrapping, setIsBootstrapping] = useState(accessToken === null)

  useEffect(() => {
    if (accessToken !== null) {
      setIsBootstrapping(false)
      return
    }

    let cancelled = false
    refreshSession()
      .then(async (token) => {
        // Bug 1.3 fix (docs bug list): populate the full user profile
        // (fullName, email) so Header can display it - previously
        // only `role` was ever recoverable after a reload, via the
        // JWT-decode stopgap in authStore.isAdmin(), even though
        // GET /auth/me exists precisely to restore the rest.
        try {
          const user = await getMe()
          if (!cancelled) useAuthStore.getState().setAuth(user, token)
        } catch {
          // /me failed but the refresh itself succeeded - stay logged
          // in with role-only info (the JWT-decode fallback still
          // covers isAdmin()); not fatal enough to log the user out.
        }
      })
      .catch(() => {
        // No valid refresh cookie - this is the normal "not logged in"
        // case, not an error. AuthGuard will redirect to /login.
      })
      .finally(() => {
        if (!cancelled) setIsBootstrapping(false)
      })

    return () => {
      cancelled = true
    }
    // Intentionally run only once on mount - accessToken changing
    // afterward (e.g. via login) should not re-trigger a bootstrap
    // refresh attempt.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return { isBootstrapping }
}
