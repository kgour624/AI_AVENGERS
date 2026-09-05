import { create } from 'zustand'
import type { User } from '@/types/auth'
import { decodeJWT } from '@/utils/jwt'

/**
 * Auth store - client state, NOT server state (no TanStack Query here).
 *
 * WHY no persist middleware: FRONTEND_SYSTEM_DESIGN.md section 15 is
 * explicit - "JWT stored in memory (Zustand) - NOT localStorage.
 * WHY: XSS cannot steal in-memory tokens." Adding zustand's `persist`
 * middleware here (even with sessionStorage) would write the access
 * token to Web Storage, defeating the entire point. On a hard refresh,
 * the access token is intentionally lost - the app must re-derive it
 * via POST /auth/refresh (httpOnly cookie), which useAuth's bootstrap
 * effect performs on mount.
 *
 * Cross-question: what happens if two components call setAuth
 * concurrently with different data (e.g. login response race with
 * a background refresh)? Zustand's set() is synchronous and last-write-
 * wins, which is correct here - a later authoritative response should
 * always win over a stale one, and axios/fetch resolve in call order
 * for a single tab, so this is not a realistic race in practice.
 */
interface AuthState {
  user: User | null
  accessToken: string | null
  setAuth: (user: User, accessToken: string) => void
  setAccessToken: (accessToken: string) => void
  clearAuth: () => void
  isAdmin: () => boolean
  isAuthenticated: () => boolean
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: null,

  setAuth: (user, accessToken) => set({ user, accessToken }),

  // WHY a separate setter: token refresh (api/base.ts interceptor) only
  // ever has a new access token, not a full User object. Forcing it
  // through setAuth would require re-fetching /me unnecessarily.
  setAccessToken: (accessToken) => set({ accessToken }),

  clearAuth: () => set({ user: null, accessToken: null }),

  // WHY fall back to decoding the JWT: see utils/jwt.ts header comment.
  // After a silent refresh (hard page reload), `user` is null but
  // `accessToken` is set - if we only checked user.role, a legitimate
  // admin would be denied access to /admin routes until some other
  // code path happened to populate `user`, which nothing currently does
  // since there is no GET /me endpoint in the documented API surface.
  isAdmin: () => {
    const { user, accessToken } = get()
    if (user) return user.role === 'admin'
    if (!accessToken) return false
    return decodeJWT(accessToken)?.role === 'admin'
  },

  isAuthenticated: () => get().accessToken !== null,
}))
