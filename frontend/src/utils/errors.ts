import axios from 'axios'
import type { ApiResponse } from '@/types/api'

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 14.
 * WHY needed now (Phase 2), not deferred: LoginPage/RegisterPage's
 * catch blocks need a single, consistent way to extract a
 * human-readable message from an axios error - without this, each
 * form would reach into `error.response.data.error.message` directly,
 * duplicating that (and getting it subtly wrong under TypeScript
 * strict mode, since `error` from a catch block is always `unknown`).
 */
export function handleAPIError(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const apiError = error.response?.data as ApiResponse<never> | undefined
    return apiError?.error?.message ?? error.message ?? 'An error occurred'
  }
  if (error instanceof Error) return error.message
  return 'Unknown error'
}
