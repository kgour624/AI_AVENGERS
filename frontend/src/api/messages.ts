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

// ============================================================
// ANSWER EXPLANATION (C9)
// ============================================================
// Everything the system knows about why an answer looks the way it does: which
// gates ran and stopped where, which sources were used, how the claims were
// verified, the quality verdict, and the integrity chain when one exists.
//
// WHY a type per section rather than one big blob: the screen must be able to say
// "this part did not run" for each section separately. A single optional bag
// would flatten "checked and clean" together with "never checked", which is the
// distinction this project keeps paying for.
export interface ExplanationMessageFacts {
  id: string
  chatId: string
  role: string
  expertId?: string
  mode: string
  gateStopped: number
  confidence?: number
}

export interface ExplanationDecisionFacts {
  route?: string
  reason?: string
  domain?: string
  genericAllowancePct?: number
}

export interface ExplanationGateStep {
  gate: number
  name?: string
  passed: boolean
  detail?: string
}

export interface ExplanationSourceRef {
  chunkId: string
  topic?: string
  sourceFile?: string
  rerankScore?: number
}

export interface ExplanationQualityFacts {
  score?: number
  verdict?: string
  method?: string
  attempts?: number
}

export interface ExplanationIntegrityFacts {
  chain?: string
  verified?: boolean
}

export interface AnswerExplanation {
  message: ExplanationMessageFacts
  decision: ExplanationDecisionFacts
  gates: ExplanationGateStep[]
  refusal?: Record<string, unknown>
  sources: ExplanationSourceRef[]
  claims?: unknown
  quality: ExplanationQualityFacts
  integrity?: ExplanationIntegrityFacts
  expert?: Record<string, unknown>
}

export const getMessageExplanation = (messageId: string) =>
  baseAPI
    .get<ApiResponse<AnswerExplanation>>(`/api/v1/messages/${messageId}/explanation`)
    .then((res) => res.data.data!)
