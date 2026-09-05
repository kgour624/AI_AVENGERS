import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/stores/authStore'
import { camelizeKeys, snakeifyKeys } from '@/utils/casing'

/**
 * Base axios instance shared by every api/*.ts file.
 * Source: FRONTEND_SYSTEM_DESIGN.md section 7, with three fixes applied
 * to the doc's illustrative snippet (each explained below because the
 * user asked for this file to survive scrutiny under multiple
 * concurrent-request scenarios, not just the happy path).
 *
 * FIX 1 - Single-flight refresh (race condition in the doc's snippet):
 * The doc's response interceptor calls `refreshToken()` independently
 * inside every failed request's own catch handler. Trace through what
 * happens if 3 requests 401 within the same tick (e.g. a page loads
 * chat + messages + experts in parallel, and the access token expired
 * a moment ago): all 3 would call POST /auth/refresh simultaneously.
 * Refresh tokens are commonly single-use/rotated server-side for
 * security - a second concurrent refresh call could get rejected by
 * the backend, or worse, invalidate the first call's new token. Fix:
 * `refreshPromise` is a module-level singleton. The first 401 starts
 * the refresh and stores the in-flight promise; subsequent 401s within
 * the same window await the SAME promise instead of starting their own.
 *
 * FIX 2 - Retry-loop guard (missing in the doc's snippet):
 * The doc's snippet unconditionally retries the original request after
 * refreshing. Trace through: if the NEW token is also rejected (e.g.
 * the user's session was revoked server-side, or clock skew causes
 * every token to look expired), the retried request 401s again, the
 * interceptor's catch block fires again, it refreshes again, retries
 * again - forever. Fix: `config._retry` flag, checked before
 * attempting a refresh; a request that has already been retried once
 * is allowed to fail through to the caller instead of looping.
 *
 * FIX 3 - Casing bridge (documented in types/api.ts):
 * Request bodies are snakeified going out, response bodies are
 * camelized coming in, so every other file in this codebase can stay
 * consistently camelCase without ever touching a network payload
 * directly.
 */

declare module 'axios' {
  export interface InternalAxiosRequestConfig {
    _retry?: boolean
  }
}

export const baseAPI = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
  // WHY withCredentials: the httpOnly refresh-token cookie (section 15,
  // 16) only gets sent by the browser on same-site/CORS requests if
  // this is true. Without it, POST /auth/refresh would silently omit
  // the cookie and always fail with "no refresh token", even though
  // the cookie exists in the browser.
  withCredentials: true,
})

// Request interceptor - attach JWT + snakeify outgoing bodies
baseAPI.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = useAuthStore.getState().accessToken
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  if (config.data && typeof config.data === 'object' && !(config.data instanceof FormData)) {
    config.data = snakeifyKeys(config.data)
  }
  return config
})

// Response interceptor - camelize incoming bodies, handle 401 via
// single-flight refresh with a one-shot retry guard.
let refreshPromise: Promise<string> | null = null

async function performRefresh(): Promise<string> {
  // Deliberately a raw axios call, NOT baseAPI - using baseAPI here
  // would re-enter this same interceptor chain if refresh itself 401s.
  const response = await axios.post<{ data: { access_token: string } }>(
    `${import.meta.env.VITE_API_URL}/api/v1/auth/refresh`,
    {},
    { withCredentials: true }
  )
  const newToken = camelizeKeys<{ accessToken: string }>(response.data.data).accessToken
  useAuthStore.getState().setAccessToken(newToken)
  return newToken
}

baseAPI.interceptors.response.use(
  (response) => {
    if (response.data && typeof response.data === 'object') {
      response.data = camelizeKeys(response.data)
    }
    return response
  },
  async (error: AxiosError) => {
    const config = error.config as InternalAxiosRequestConfig | undefined

    const isUnauthorized = error.response?.status === 401
    const alreadyRetried = config?._retry === true
    const isRefreshCall = config?.url?.includes('/auth/refresh')

    if (!isUnauthorized || alreadyRetried || !config || isRefreshCall) {
      // Not a recoverable 401, already tried once, or this WAS the
      // refresh call failing - propagate as-is in every case.
      return Promise.reject(error)
    }

    config._retry = true

    try {
      // Single-flight: only the first caller in a burst actually hits
      // the network; everyone else awaits the same promise.
      if (!refreshPromise) {
        refreshPromise = performRefresh().finally(() => {
          refreshPromise = null
        })
      }
      const newToken = await refreshPromise

      config.headers = config.headers ?? {}
      config.headers.Authorization = `Bearer ${newToken}`
      return baseAPI(config)
    } catch (refreshError) {
      useAuthStore.getState().clearAuth()
      if (typeof window !== 'undefined') {
        window.location.href = '/login'
      }
      return Promise.reject(refreshError)
    }
  }
)
