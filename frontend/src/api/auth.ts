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
