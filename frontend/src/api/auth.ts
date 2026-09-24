import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'
import type { AdminLoginRequest, LoginRequest, RegisterRequest, TokenPair, User } from '@/types/auth'

export const login = (req: LoginRequest) =>
  baseAPI
    .post<ApiResponse<{ user: User; tokenPair: TokenPair }>>('/api/v1/auth/login', req)
    .then((res) => res.data.data!)

export const register = (req: RegisterRequest) =>
  baseAPI
    .post<ApiResponse<{ user: User; tokenPair: TokenPair }>>('/api/v1/auth/register', req)
    .then((res) => res.data.data!)

export const adminLogin = (req: AdminLoginRequest) =>
  baseAPI
    .post<ApiResponse<{ user: User; tokenPair: TokenPair }>>('/api/v1/auth/admin/login', {
      email: req.email,
      password: req.password,
      totp_code: req.totpCode,
    })
    .then((res) => res.data.data!)

export const logout = () => baseAPI.post('/api/v1/auth/logout').then((res) => res.data)

// Bug 1.3 fix (docs bug list): GET /auth/me exists on the backend
// specifically to restore the full profile after a hard reload
// (HANDOFF.md Gap #2), but nothing on the frontend ever called it -
// authStore.user stayed null forever after a reload, so Header had
// no name/email to show even once this endpoint existed. Wired into
// hooks/useAuth.ts's bootstrap flow.
export const getMe = () => baseAPI.get<ApiResponse<User>>('/api/v1/auth/me').then((res) => res.data.data!)

// forgotPassword — POST /auth/forgot-password
// Always resolves (backend never reveals whether the email exists).
export const forgotPassword = (email: string) =>
  baseAPI
    .post<ApiResponse<{ message: string }>>('/api/v1/auth/forgot-password', { email })
    .then((res) => res.data.data!)

// resetPassword — POST /auth/reset-password
// token: the raw token from the reset link (?token=...)
// newPassword: must be ≥8 characters
export const resetPassword = (token: string, newPassword: string) =>
  baseAPI
    .post<ApiResponse<{ message: string }>>('/api/v1/auth/reset-password', {
      token,
      new_password: newPassword,
    })
    .then((res) => res.data.data!)

// Hidden admin bootstrap (unguessable token; not linked from login/register)
export const bootstrapAdminStart = (req: {
  token: string
  email: string
  password: string
  fullName: string
}) =>
  baseAPI
    .post<ApiResponse<{ secret: string; qrUrl: string }>>('/api/v1/auth/admin/bootstrap/start', {
      token: req.token,
      email: req.email,
      password: req.password,
      full_name: req.fullName,
    })
    .then((res) => res.data.data!)

export const bootstrapAdminComplete = (req: { token: string; totpCode: string }) =>
  baseAPI
    .post<ApiResponse<{ user: User; tokenPair: TokenPair }>>('/api/v1/auth/admin/bootstrap/complete', {
      token: req.token,
      totp_code: req.totpCode,
    })
    .then((res) => res.data.data!)

export const getTotpStatus = () =>
  baseAPI.get<ApiResponse<{ totpEnabled: boolean }>>('/api/v1/auth/totp').then((res) => res.data.data!)

export const setupTotp = () =>
  baseAPI
    .post<ApiResponse<{ secret: string; qrUrl: string }>>('/api/v1/auth/totp/setup')
    .then((res) => res.data.data!)

export const enableTotp = (code: string) =>
  baseAPI
    .post<ApiResponse<{ status: string }>>('/api/v1/auth/totp/enable', { code })
    .then((res) => res.data.data!)

export const disableTotp = () =>
  baseAPI
    .post<ApiResponse<{ status: string }>>('/api/v1/auth/totp/disable')
    .then((res) => res.data.data!)
