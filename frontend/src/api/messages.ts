import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'

/**
 * NOTE: sendMessage is intentionally NOT here. Per
 * FRONTEND_SYSTEM_DESIGN.md section 7 comment: "sendMessage uses SSE -
 * NOT axios." It lives in hooks/useSSEStream.ts because streaming
 * responses need raw fetch() + ReadableStream, which axios does not
 * support natively. Only the (non-streaming) rating endpoint belongs
 * in this axios-based file.
 *
 * PHASE 4 CORRECTION: RateRequest below was incomplete - verified
 * against the real backend-go/internal/rating/handler.go RateRequest
 * struct, which also has codeExecuted/executionSuccess/errorMessage
 * fields (used when a code suggestion was actually run by the client
 * and failed/succeeded). Added them as optional so RatingWidget (which
 * only ever sends score today) keeps compiling unchanged, but a future
 * "code execution feedback" feature has real fields to send against.
 */
export interface RateRequest {
  score: 1 | 2 | 3 | 4 | 5
  feedback?: string
  feedbackType?: 'accepted' | 'rejected' | 'modified' | 'ignored'
  codeExecuted?: boolean
  executionSuccess?: boolean
  errorMessage?: string
}

export const rateMessage = (messageId: string, rating: RateRequest) =>
  baseAPI.post(`/api/v1/messages/${messageId}/rate`, rating).then((res) => res.data)

export const deleteMessage = (messageId: string) =>
  baseAPI
    .delete<ApiResponse<{ status: string }>>(`/api/v1/messages/${messageId}`)
    .then((res) => res.data.data!)

export const updateMessage = (messageId: string, content: string) =>
  baseAPI
    .patch<ApiResponse<{ status: string }>>(`/api/v1/messages/${messageId}`, { content })
    .then((res) => res.data.data!)
