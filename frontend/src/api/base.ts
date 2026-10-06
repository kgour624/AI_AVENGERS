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
// FIX api call increase: token ko store + localStorage dono se padho, header case normalize, FormData boundary fix
baseAPI.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const storeToken = useAuthStore.getState().accessToken
  const lsToken = (typeof localStorage !== 'undefined' && (localStorage.getItem('accessToken') || localStorage.getItem('access_token') || localStorage.getItem('token'))) || null
  const token = storeToken || lsToken
  if (token) {
    if (config.headers && typeof (config.headers as any).set === 'function') {
      (config.headers as any).set('Authorization', `Bearer ${token}`)
    } else {
      config.headers = config.headers || ({} as any)
      ;(config.headers as any)['Authorization'] = `Bearer ${token}`
    }
  }
  if (config.data && typeof config.data === 'object' && !(config.data instanceof FormData)) {
    config.data = snakeifyKeys(config.data)
  }
  if (config.data instanceof FormData) {
    // Browser ko boundary set karne do - explicit Content-Type hatao
    // Robust: AxiosHeaders instance (Axios v1) vs plain object dono handle
    try {
      const h: any = config.headers
      if (h) {
        if (typeof h.delete === 'function') {
          try { h.delete('Content-Type') } catch {}
          try { h.delete('content-type') } catch {}
          try { h.delete('Content-type') } catch {}
        } else {
          try { delete h['Content-Type'] } catch {}
          try { delete h['content-type'] } catch {}
          try { delete h['Content-type'] } catch {}
          try { delete h['common']?.['Content-Type'] } catch {}
        }
        // Ensure axios doesn't re-add json header - set undefined so it is omitted
        try { h['Content-Type'] = undefined } catch {}
      }
      // Also scrub common default for this single request
      if (config.headers && typeof (config.headers as any).set === 'function') {
        try { (config.headers as any).set('Content-Type', undefined as any) } catch {}
      }
    } catch {}
  }
  return config
})

// Response interceptor - camelize incoming bodies, handle 401 via
// single-flight refresh with a one-shot retry guard.
let refreshPromise: Promise<string> | null = null

// Exported (not just used internally by the interceptor below) because
// useAuth's bootstrap flow (src/hooks/useAuth.ts) also needs to attempt
// a silent refresh on app mount, before any request has even 401'd -
// e.g. immediately after a hard page reload when accessToken is null.
export async function refreshSession(): Promise<string> {
  // Deliberately a raw axios call, NOT baseAPI.
  // WHY: using baseAPI would re-enter this interceptor chain if refresh
  // itself 401s, causing an infinite loop.
  //
  // Body is intentionally empty {}.
  // WHY: refresh token travels as an httpOnly cookie (set by backend
  // on login). withCredentials:true makes the browser send it
  // automatically. JavaScript never touches the refresh token.
  //
  // Response shape: { success, data: { accessToken, expiresInSeconds }, meta }
  // (refresh token is NOT in response body — it's rotated in the cookie)
  //
  // WHY camelCase here now: backend-go/cmd/server/main.go's handleRefresh
  // was updated in this same commit to emit camelCase directly (Interface-
  // First audit, docs/INTERFACE_FIRST_CONTRACT.md §4). This call bypasses
  // baseAPI's camelizeKeys() response interceptor on purpose (raw axios,
  // to avoid interceptor re-entrancy if refresh itself 401s), so it reads
  // the wire shape directly - it must be updated in lockstep with the
  // backend handler, not left to drift.
  const res = await axios.post<any>(
    `${import.meta.env.VITE_API_URL}/api/v1/auth/refresh`,
    {},
    { withCredentials: true }
  )
  // FUTURE-PROOF extractor: backend envelope may be {success,data:{accessToken}} or {data:{access_token}} or flat
  // Screenshot bug: refresh 200 but retry 401 -> old code did res.data.data.accessToken only -> undefined on snake_case -> retry with undefined
  const d: any = res.data
  const newToken: string =
    d?.data?.accessToken ||
    d?.data?.access_token ||
    d?.accessToken ||
    d?.access_token ||
    d?.data?.token ||
    d?.token ||
    ''
  if (!newToken) throw new Error('No access token in refresh response')
  useAuthStore.getState().setAccessToken(newToken)
  try {
    localStorage.setItem('access_token', newToken)
    localStorage.setItem('accessToken', newToken)
    localStorage.setItem('token', newToken)
  } catch {}
  baseAPI.defaults.headers.common['Authorization'] = `Bearer ${newToken}`
  return newToken
}

// FIX api call increase: proper queue + circuit breaker. Pehle har 401 alag refresh marta tha -> N calls. Ab 1 hi refresh + queue.
let failedQueue: Array<{ resolve: (token: string) => void; reject: (err: any) => void }> = []
function flushQueue(err: any, token: string | null) {
  failedQueue.forEach((p) => (token ? p.resolve(token) : p.reject(err)))
  failedQueue = []
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
    const isAuthCall = config?.url?.includes('/auth/login') || config?.url?.includes('/auth/register') || config?.url?.includes('/auth/admin/login')

    if (!isUnauthorized || alreadyRetried || !config || isRefreshCall || isAuthCall) {
      if (isUnauthorized && alreadyRetried) {
        // Retried request still 401'd -> token is invalid or unauthorized even after refresh
        try { useAuthStore.getState().clearAuth() } catch {}
        try {
          localStorage.removeItem('accessToken')
          localStorage.removeItem('access_token')
          localStorage.removeItem('token')
        } catch {}
        if (typeof window !== 'undefined') {
          const cur = window.location.pathname
          if (cur !== '/login' && cur !== '/admin/login') window.location.href = '/login'
        }
      }
      return Promise.reject(error)
    }

    config._retry = true

    // already refreshing -> queue this request, dont fire another refresh
    if (refreshPromise) {
      return new Promise<string>((resolve, reject) => {
        failedQueue.push({ resolve, reject })
      })
        .then((newToken) => {
          if (config.headers && typeof (config.headers as any).set === 'function') {
            (config.headers as any).set('Authorization', `Bearer ${newToken}`)
          } else {
            config.headers = config.headers || ({} as any)
            ;(config.headers as any)['Authorization'] = `Bearer ${newToken}`
          }
          baseAPI.defaults.headers.common['Authorization'] = `Bearer ${newToken}`
          return baseAPI(config)
        })
        .catch((e) => Promise.reject(e))
    }

    // first 401 -> start refresh
    refreshPromise = refreshSession()
      .then((t) => {
        flushQueue(null, t)
        return t
      })
      .catch((e) => {
        flushQueue(e, null)
        throw e
      })
      .finally(() => {
        refreshPromise = null
      })

    try {
      const newToken = await refreshPromise
      if (!newToken) throw new Error('Refresh returned empty token')
      if (config.headers && typeof (config.headers as any).set === 'function') {
        (config.headers as any).set('Authorization', `Bearer ${newToken}`)
      } else {
        config.headers = config.headers || ({} as any)
        ;(config.headers as any)['Authorization'] = `Bearer ${newToken}`
      }
      baseAPI.defaults.headers.common['Authorization'] = `Bearer ${newToken}`
      return baseAPI(config)
    } catch (refreshError) {
      try { useAuthStore.getState().clearAuth() } catch {}
      try {
        localStorage.removeItem('accessToken')
        localStorage.removeItem('access_token')
        localStorage.removeItem('token')
      } catch {}
      if (typeof window !== 'undefined') {
        // sirf ek baar redirect, loop nahi
        const cur = window.location.pathname
        if (cur !== '/login' && cur !== '/admin/login') window.location.href = '/login'
      }
      return Promise.reject(refreshError)
    }
  }
)
