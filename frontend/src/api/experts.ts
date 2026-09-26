import { baseAPI } from './base'
import type { ApiResponse } from '@/types/api'
import type { Expert, ExpertTopic } from '@/types/expert'

export const getExperts = (options?: { signal?: AbortSignal }) =>
  baseAPI
    .get<ApiResponse<Expert[]>>('/api/v1/experts', { signal: options?.signal })
    .then((res) => res.data.data!)

export const getExpert = (id: string, options?: { signal?: AbortSignal }) =>
  baseAPI
    .get<ApiResponse<Expert>>(`/api/v1/experts/${id}`, { signal: options?.signal })
    .then((res) => res.data.data!)

export const getExpertTopics = (id: string, options?: { signal?: AbortSignal }) =>
  baseAPI
    .get<ApiResponse<{ topics: ExpertTopic[] }>>(`/api/v1/experts/${id}/topics`, {
      signal: options?.signal,
    })
    .then((res) => res.data.data!.topics)

/**
 * T-CAT: the named answer formats available for an expert's category.
 * Empty `formats` means the expert has no named variants (flat text or the
 * category's single template) — the chat shows no selector in that case.
 */
export interface ExpertAnswerFormats {
  category: string
  default: string
  formats: string[]
}

export const getExpertAnswerFormats = (id: string, options?: { signal?: AbortSignal }) =>
  baseAPI
    .get<ApiResponse<ExpertAnswerFormats>>(`/api/v1/experts/${id}/answer-formats`, {
      signal: options?.signal,
    })
    .then((res) => res.data.data ?? { category: '', default: '', formats: [] })
