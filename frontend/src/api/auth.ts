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
    .post<ApiResponse<{ user: User; tokenPair: TokenPair }>>('/api/v1/auth/admin/login', req)
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
