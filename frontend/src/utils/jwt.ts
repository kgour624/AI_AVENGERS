/**
 * Client-side JWT claim decoding (NOT verification - verifying the
 * signature is the backend's job; the frontend only ever trusts a
 * token because the browser received it over HTTPS from our own API).
 *
 * WHY this file exists - a gap in both design docs:
 * AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 15 lists the full auth
 * API surface (register, login, admin/login, refresh, logout) and
 * there is no GET /me or /users/me endpoint anywhere in either doc.
 * Trace through what that means on a hard page refresh: accessToken
 * is intentionally NOT persisted (see authStore.ts), so after a
 * refresh the app has no token and no user object. It calls
 * POST /auth/refresh (httpOnly cookie) to get a new access token -
 * but that response only contains the token, not who it belongs to.
 * Without decoding the JWT, AuthGuard would have no way to decide
 * "is this session valid" or "is this user an admin" post-refresh,
 * and the /admin routes' guard would incorrectly deny access to a
 * legitimately re-authenticated admin.
 *
 * This is a stopgap, not a real fix - it only recovers `role` and
 * `sub` (user id) from the token payload, not fullName/email/etc.
 * The header would show no display name until the user navigates
 * somewhere that fetches their profile some other way. The correct
 * fix is for the backend to add GET /api/v1/me - flagged in
 * HANDOFF.md as a backend gap, not silently worked around forever.
 */

export interface DecodedJWTClaims {
  sub: string
  email?: string
  role?: 'admin' | 'client'
  exp: number
}

export function decodeJWT(token: string): DecodedJWTClaims | null {
  try {
    const payload = token.split('.')[1]
    if (!payload) return null
    const base64 = payload.replace(/-/g, '+').replace(/_/g, '/')
    const json = atob(base64)
    return JSON.parse(json) as DecodedJWTClaims
  } catch {
    return null
  }
}

export function isTokenExpired(token: string): boolean {
  const claims = decodeJWT(token)
  if (!claims) return true
  return claims.exp * 1000 < Date.now()
}
