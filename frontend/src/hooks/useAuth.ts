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

    // WHY tryRefresh helper:
    //   On page reload, the browser may take a moment to send the
    //   httpOnly cookie on a cross-port request (3001 → 8080).
    //   We retry once after 800ms before giving up.
    //   This prevents spurious logouts on reload without masking
    //   genuine "not logged in" cases.
    const tryRefresh = async (attempt: number): Promise<string> => {
      try {
        return await refreshSession()
      } catch (err) {
        if (attempt === 0 && !cancelled) {
          // First attempt failed — wait 800ms and retry once
          await new Promise((resolve) => setTimeout(resolve, 800))
          if (!cancelled) return tryRefresh(1)
        }
        throw err
      }
    }

    tryRefresh(0)
      .then(async (token) => {
        try {
          const user = await getMe()
          if (!cancelled) useAuthStore.getState().setAuth(user, token)
        } catch {
          // /me failed but refresh succeeded — stay logged in with
          // role-only info from JWT decode fallback.
        }
      })
      .catch(() => {
        // No valid refresh cookie after retry — genuinely not logged in.
        // AuthGuard will redirect to /login.
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
