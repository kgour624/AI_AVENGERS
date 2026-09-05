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
