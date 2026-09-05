import { baseAPI } from './base'

/**
 * NOTE: sendMessage is intentionally NOT here. Per
 * FRONTEND_SYSTEM_DESIGN.md section 7 comment: "sendMessage uses SSE -
 * NOT axios." It lives in hooks/useSSEStream.ts because streaming
 * responses need raw fetch() + ReadableStream, which axios does not
 * support natively. Only the (non-streaming) rating endpoint belongs
 * in this axios-based file.
 */

export interface RateRequest {
  score: 1 | 2 | 3 | 4 | 5
  feedback?: string
  feedbackType?: 'accepted' | 'rejected' | 'modified' | 'ignored'
}

export const rateMessage = (messageId: string, rating: RateRequest) =>
  baseAPI.post(`/api/v1/messages/${messageId}/rate`, rating).then((res) => res.data)
