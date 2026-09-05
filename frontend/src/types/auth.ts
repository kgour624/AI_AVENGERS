/**
 * Auth domain models.
 * Source: FRONTEND_SYSTEM_DESIGN.md sections 6, 13, 15.
 *
 * SECURITY NOTE: There is intentionally NO `refreshToken` field on
 * any type here. Per section 15/16 of both design docs, the refresh
 * token lives in an httpOnly cookie set by the backend - JavaScript
 * must never be able to read it (that's what makes it XSS-resistant).
 * If a future change adds a refreshToken field visible to the client,
 * that defeats the entire point of the httpOnly-cookie design and
 * should be treated as a regression, not a feature.
 */

export type UserRole = 'admin' | 'client'

export interface User {
  id: string
  email: string
  fullName: string
  role: UserRole
  totpEnabled: boolean
}

/** Only the access token is ever held in JS memory (Zustand authStore). */
export interface TokenPair {
  accessToken: string
  expiresInSeconds: number
}

export interface LoginRequest {
  email: string
  password: string
}

export interface AdminLoginRequest extends LoginRequest {
  totpCode: string
}

export interface RegisterRequest {
  email: string
  password: string
  fullName: string
}
