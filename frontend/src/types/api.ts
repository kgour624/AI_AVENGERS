/**
 * API response envelope + shared primitives.
 *
 * WHY this file exists separately from resource-specific type files:
 * APIResponse<T> wraps every endpoint's response - defining it once
 * avoids duplicating the envelope shape across expert.ts/project.ts/etc.
 *
 * NOTE ON CASING (gap found in FRONTEND_SYSTEM_DESIGN.md section 13 vs
 * AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 15):
 * The backend's example JSON response uses snake_case (`request_id`),
 * but every TypeScript interface in the design doc is camelCase
 * (`requestId`). The doc never shows a conversion step. If left
 * unhandled, `response.data.meta.requestId` would be `undefined` at
 * runtime (the real key is `request_id`) while TypeScript happily
 * compiles it - a silent bug that would only surface when someone
 * tries to read the field. See src/utils/casing.ts for the fix: the
 * axios response interceptor in api/base.ts deep-camelizes every
 * response body before it reaches application code, so the camelCase
 * types below are honored end to end. This decision is logged in
 * HANDOFF.md.
 */

export interface ApiErrorShape {
  code: string
  message: string
  details?: Record<string, string>
}

export interface ApiMeta {
  requestId: string
  timestamp: string
}

export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: ApiErrorShape
  meta: ApiMeta
}

/** Pagination shape used by list endpoints (timeline, violations, etc.) */
export interface PaginatedResponse<T> {
  items: T[]
  nextCursor?: string
  hasMore: boolean
}
