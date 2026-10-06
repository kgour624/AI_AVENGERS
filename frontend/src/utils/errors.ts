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
    const backendMsg = (apiError?.error?.message || (apiError as unknown as { message?: string })?.message) as string | undefined
    if (backendMsg && typeof backendMsg === 'string' && backendMsg.trim()) {
      const lower = backendMsg.toLowerCase()
      if (lower.includes('unsupported_format') || lower.includes('unsupported format')) return 'This file type is not supported. Please upload PDF, Word, Excel, PPT, CSV or text/markdown files.'
      if (lower.includes('file_required') || lower.includes('no file')) return 'No file was received. Please select a file and try again.'
      if (lower.includes('payload too large') || lower.includes('too large')) return backendMsg
      return backendMsg
    }
    if (!error.response) {
      const msg = error.message || ''
      const lower = msg.toLowerCase()
      if (lower.includes('config.headers.delete') || lower.includes('is not a function')) return 'Upload failed due to a browser issue. Please refresh the page and try again.'
      if (lower.includes('network error') || lower.includes('failed to fetch') || lower.includes('networkerror')) return 'Network error — please check your internet and try again.'
      if (lower.includes('timeout')) return 'Request timed out — server is busy. Please try again.'
      return msg || 'Network error — please try again.'
    }
    const status = error.response.status
    if (status === 413) return 'File is too large. Maximum size is 50MB.'
    if (status === 401) return 'Session expired. Please login again.'
    if (status === 403) return 'You do not have permission to do this.'
    if (status === 429) return 'Too many requests. Please wait a moment and retry.'
    if (status >= 500) return 'Server error — please try again in a moment.'
    return error.message || 'Something went wrong. Please try again.'
  }
  if (error instanceof Error) {
    const lower = error.message.toLowerCase()
    if (lower.includes('config.headers.delete') || lower.includes('is not a function')) return 'Upload failed due to a browser issue. Please refresh and try again.'
    return error.message || 'Something went wrong.'
  }
  return 'Unknown error — please try again.'
}
